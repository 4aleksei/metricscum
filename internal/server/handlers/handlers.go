// Package handlers - servers endpoint realization
package handlers

import (
	"bytes"
	"database/sql"

	"crypto/rsa"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/4aleksei/metricscum/internal/common/middleware/hmacsha256"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/handlers/middleware/httpaes"
	"github.com/4aleksei/metricscum/internal/server/handlers/middleware/httpgzip"
	"github.com/4aleksei/metricscum/internal/server/handlers/middleware/httphmacsha256"
	"github.com/4aleksei/metricscum/internal/server/handlers/middleware/httplogs"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type (
	HandlersServer struct {
		store       *service.HandlerStore
		cfg         *config.Config
		Srv         *http.Server
		l           *zap.Logger
		key         string
		privateKey  *rsa.PrivateKey
		trustedCidr *net.IPNet
	}
)

const (
	textHTMLContent        string = "text/html"
	applicationJSONContent string = "application/json"
)

// NewServer - server constructor
// store : store object
// cfg : config
// l :  logger realization
func NewServer(store *service.HandlerStore, cfg *config.Config, l *zap.Logger) (*HandlersServer, error) {
	h := new(HandlersServer)
	h.store = store
	h.cfg = cfg
	h.key = h.cfg.Key
	h.l = l

	if cfg.Cidr != "" {
		var err error
		_, h.trustedCidr, err = net.ParseCIDR(cfg.Cidr)
		if err != nil {
			return nil, err
		}
	}

	if h.cfg.PrivateKeyFile != "" {
		pKey, err := httpaes.LoadKey(h.cfg.PrivateKeyFile)
		if err != nil {
			h.l.Debug("HTTP server Load private key error: ", zap.Error(err))
			return nil, err
		} else {
			h.privateKey = pKey
		}
	}

	h.Srv = &http.Server{
		Addr:              h.cfg.Address,
		Handler:           h.newRouter(),
		ReadHeaderTimeout: 2 * time.Second,
	}
	return h, nil
}

func (h *HandlersServer) withLogging(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := httplogs.NewResponseData()
		lw := httplogs.NewResponseWriter(responseData, w)

		next.ServeHTTP(lw, r)
		duration := time.Since(start)
		h.l.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("AcceptEnc", r.Header.Get("Accept-Encoding")),
			zap.String("ContentEnc", r.Header.Get("Content-Encoding")),
			zap.String("Accept", r.Header.Get("Accept")),
			zap.String("AES-256", r.Header.Get("AES-256")),
			zap.String("X-Real-IP", r.Header.Get("X-Real-IP")),
			zap.String("ContentType", r.Header.Get("Content-Type")),
			zap.Duration("duration", duration),
			zap.Int("resp_status", responseData.GetStatus()),
			zap.Int("resp_size", responseData.GetSize()))
	}
	return http.HandlerFunc(logFn)
}

// Serve - start server in go-routine
func (h *HandlersServer) Serve() {
	go func() {
		if err := h.Srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			h.l.Debug("HTTP server error: ", zap.Error(err))
		}
		h.l.Info("Stopped serving new connections.")
	}()
}

func (h *HandlersServer) aesMiddleware(next http.Handler) http.Handler {
	aesfn := func(w http.ResponseWriter, r *http.Request) {
		aesEncoding := r.Header.Get("AES-256")
		if aesEncoding != "" {
			ar, err := httpaes.NewAesReader(r.Body, h.privateKey, aesEncoding)
			if err != nil {
				h.l.Debug("cannot decode aes", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = ar
			defer ar.Close()
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(aesfn)
}

func (h *HandlersServer) gzipMiddleware(next http.Handler) http.Handler {
	gzipfn := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := httpgzip.NewCompressWriter(w)
			ow = cw
			defer cw.Close()
		}
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := httpgzip.NewCompressReader(r.Body)
			if err != nil {
				h.l.Debug("cannot decode gzip", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(gzipfn)
}

func (h *HandlersServer) hmacsha256Middleware(next http.Handler) http.Handler {
	hmacsha256fn := func(w http.ResponseWriter, r *http.Request) {
		ow := httphmacsha256.NewWriter(w, []byte(h.cfg.Key))

		r.Body = hmacsha256.NewReader(r.Body, []byte(h.cfg.Key))

		next.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(hmacsha256fn)
}

func (h *HandlersServer) trustedCIDRMiddleware(next http.Handler) http.Handler {
	checkfn := func(w http.ResponseWriter, r *http.Request) {

		realIP := r.Header.Get("X-Real-IP")
		if realIP == "" {
			realIP = r.RemoteAddr
		}

		ip := net.ParseIP(realIP)

		if ip == nil {
			h.l.Debug("cannot parse ip address")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		con := h.trustedCidr.Contains(ip)
		if !con {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(checkfn)
}

func (h *HandlersServer) newRouter() http.Handler {
	mux := chi.NewRouter()

	mux.Use(h.withLogging)


	if h.trustedCidr != nil {
		mux.Use(h.trustedCIDRMiddleware)
	}

	if h.privateKey != nil {
		mux.Use(h.aesMiddleware)
	}

	mux.Use(h.gzipMiddleware)
	if h.key != "" {
		mux.Use(h.hmacsha256Middleware)
	}

	mux.Use(middleware.Recoverer)

	mux.Mount("/debug", middleware.Profiler())

	mux.Post("/update/", h.mainPageJSON)
	mux.Post("/updates/", h.mainPageJSONs)
	mux.Post("/update/{type}/{name}/{value}", h.mainPostPagePlain)
	mux.Post("/update/{type}/", h.mainPageFoundErrors)
	mux.Post("/*", h.mainPageError)
	mux.Get("/value/{type}/{name}", h.mainPageGetPlain)
	mux.Post("/value/", h.mainPageGetJSON)
	mux.Get("/ping", h.mainPingDB)
	mux.Get("/", h.mainPage)

	return mux
}

func (h *HandlersServer) checkHmacSha256(res http.ResponseWriter, req *http.Request) bool {
	if h.key != "" {
		sig, err := hmacsha256.GetSig(req.Body)
		if err != nil {
			h.l.Error("error read request", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return false
		}

		sigBody := req.Header.Get("HashSHA256")
		if sigBody == "" { // accept data if no hash in body, but with secret key in app parameters . strange
			return true
		}
		sigString := hex.EncodeToString(sig)
		if sigString != sigBody {
			h.l.Debug("Signature in body NOT equal calculated signature", zap.String("b", sigBody), zap.String("cb", sigString))
			http.Error(res, "Bad request!", http.StatusBadRequest)
			return false
		} else {
			h.l.Debug("Signature in body accepted")
		}
	}
	return true
}

// едпоинт  POST /update/
func (h *HandlersServer) mainPageJSON(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != applicationJSONContent {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}
	var JSONstr models.Metrics
	if err := JSONstr.JSONDecode(req.Body); err != nil {
		h.l.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !h.checkHmacSha256(res, req) {
		return
	}

	val, err := h.store.SetValueModel(req.Context(), JSONstr)
	if err != nil {
		if errors.Is(err, service.ErrBadName) {
			http.Error(res, "Invalid request!", http.StatusNotFound)
			return
		}
		http.Error(res, "Invalid request!", http.StatusBadRequest)
		return
	}
	var buf bytes.Buffer
	if errson := val.JSONEncodeBytes(io.Writer(&buf)); errson != nil {
		h.l.Debug("error encoding response", zap.Error(errson))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", applicationJSONContent)
	res.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(res, buf.String()); err != nil {
		h.l.Debug("error writing response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HandlersServer) mainPageJSONs(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != applicationJSONContent {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	JSONstrs, err := models.JSONSDecode(req.Body)

	if err != nil {
		h.l.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !h.checkHmacSha256(res, req) {
		return
	}

	val, err := h.store.SetValueSModel(req.Context(), JSONstrs)
	if err != nil {
		if errors.Is(err, service.ErrBadName) {
			http.Error(res, "Invalid request!", http.StatusNotFound)
			return
		}
		http.Error(res, "Invalid request!", http.StatusBadRequest)
		return
	}
	var buf bytes.Buffer
	if errson := models.JSONSEncodeBytes(io.Writer(&buf), val); errson != nil {
		h.l.Debug("error encoding response", zap.Error(errson))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", applicationJSONContent)
	res.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(res, buf.String()); err != nil {
		h.l.Debug("error writing response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HandlersServer) mainPageGetJSON(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != applicationJSONContent {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}
	var JSONstr models.Metrics
	if err := JSONstr.JSONDecode(req.Body); err != nil {
		h.l.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !h.checkHmacSha256(res, req) {
		return
	}

	val, err := h.store.GetValueModel(req.Context(), JSONstr)
	if err != nil {
		if errors.Is(err, service.ErrBadName) || errors.Is(err, memstorage.ErrNotFoundName) || errors.Is(err, sql.ErrNoRows) {
			http.Error(res, "Not found!", http.StatusNotFound)
			return
		}
		h.l.Debug("cannot decode GetValueModel request JSON body", zap.Error(err))
		http.Error(res, "Invalid request!", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	errson := val.JSONEncodeBytes(io.Writer(&buf))
	if errson != nil {
		h.l.Debug("error encoding response", zap.Error(errson))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", applicationJSONContent)
	res.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(res, buf.String()); err != nil {
		h.l.Debug("error writing response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HandlersServer) mainPageError(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Bad request", http.StatusBadRequest)
}

func (h *HandlersServer) mainPageFoundErrors(res http.ResponseWriter, req *http.Request) {
	typeVal := chi.URLParam(req, "type")
	errKind := h.store.CheckType(typeVal)
	if errKind != nil {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}
	http.Error(res, "Not Found", http.StatusNotFound)
}

func (h *HandlersServer) mainPostPagePlain(res http.ResponseWriter, req *http.Request) {
	typeVal := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")
	if typeVal == "" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}
	if name == "" {
		http.Error(res, "Bad name!", http.StatusNotFound)
		return
	}
	if value == "" {
		http.Error(res, "Bad data!", http.StatusBadRequest)
		return
	}
	err := h.store.RecievePlainValue(req.Context(), typeVal, name, value)
	if err != nil {
		http.Error(res, "Bad value!", http.StatusBadRequest)
		return
	}
	switch req.Header.Get("Accept") {
	case textHTMLContent:
		res.Header().Add("Content-Type", textHTMLContent)
	case applicationJSONContent:
		res.Header().Add("Content-Type", applicationJSONContent)
	default:
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	}
	res.WriteHeader(http.StatusOK)
}

func (h *HandlersServer) mainPageGetPlain(res http.ResponseWriter, req *http.Request) {
	typeVal := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	if typeVal == "" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}
	if name == "" {
		http.Error(res, "Bad type!", http.StatusNotFound)
		return
	}
	val, err := h.store.GetValuePlain(req.Context(), name, typeVal)
	if err != nil {
		h.l.Debug("error get val", zap.Error(err))
		http.Error(res, "Not found value!", http.StatusNotFound)
		return
	}
	switch req.Header.Get("Accept") {
	case textHTMLContent:
		res.Header().Add("Content-Type", textHTMLContent)
	case applicationJSONContent:
		res.Header().Add("Content-Type", applicationJSONContent)
	default:
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	}
	res.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(res, val); err != nil {
		h.l.Debug("error writing response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HandlersServer) mainPingDB(res http.ResponseWriter, req *http.Request) {
	err := h.store.GetPingDB(req.Context())
	if err != nil {
		h.l.Debug("error ping db", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch req.Header.Get("Accept") {
	case textHTMLContent:
		res.Header().Add("Content-Type", textHTMLContent)
	case applicationJSONContent:
		res.Header().Add("Content-Type", applicationJSONContent)
	default:
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	}
	res.WriteHeader(http.StatusOK)
}

func (h *HandlersServer) mainPage(res http.ResponseWriter, req *http.Request) {
	if req.URL.String() == "" || req.URL.String() == "/" {
		val, err := h.store.GetAllStore(req.Context())
		if err != nil {
			http.Error(res, "Not found value!", http.StatusNotFound)
			return
		}
		switch req.Header.Get("Accept") {
		case textHTMLContent:
			res.Header().Add("Content-Type", textHTMLContent)
		case applicationJSONContent:
			res.Header().Add("Content-Type", applicationJSONContent)
		default:
			res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		}
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write([]byte(val)); err != nil {
			h.l.Debug("error writing response", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(res, "Bad request", http.StatusBadRequest)
	}
}
