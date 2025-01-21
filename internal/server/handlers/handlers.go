package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/handlers/middleware/httpgzip"
	"github.com/4aleksei/metricscum/internal/server/handlers/middleware/httplogs"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type (
	HandlersServer struct {
		store *service.HandlerStore
		cfg   *config.Config
		Srv   *http.Server
		l     *zap.Logger
	}
)

const (
	textHTMLContent        string = "text/html"
	applicationJSONContent string = "application/json"
)

func NewHandlers(store *service.HandlerStore, cfg *config.Config, l *zap.Logger) *HandlersServer {
	h := new(HandlersServer)
	h.store = store
	h.cfg = cfg
	h.l = l
	h.Srv = &http.Server{
		Addr:              h.cfg.Address,
		Handler:           h.newRouter(),
		ReadHeaderTimeout: 2 * time.Second,
	}
	return h
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
			zap.String("ContentType", r.Header.Get("Content-Type")),
			zap.Duration("duration", duration),
			zap.Int("resp_status", responseData.GetStatus()),
			zap.Int("resp_size", responseData.GetSize()))
	}
	return http.HandlerFunc(logFn)
}

func (h *HandlersServer) Serve() {

	go func() {
		if err := h.Srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			h.l.Debug("HTTP server error: ", zap.Error(err))
		}
		h.l.Info("Stopped serving new connections.")
	}()

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

func (h *HandlersServer) newRouter() http.Handler {
	mux := chi.NewRouter()

	mux.Use(h.withLogging)
	mux.Use(h.gzipMiddleware)
	mux.Use(middleware.Recoverer)

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
