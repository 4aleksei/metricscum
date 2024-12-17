package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type HandlersServer struct {
	store *service.HandlerStore
	cfg   *config.Config
	srv   *http.Server
}

func NewHandlers(store *service.HandlerStore, cfg *config.Config) *HandlersServer {
	h := new(HandlersServer)
	h.store = store
	h.cfg = cfg
	h.srv = &http.Server{
		Addr:    h.cfg.Address,
		Handler: h.newRouter(),
	}
	return h
}

func (h *HandlersServer) Serve() error {
	return h.srv.ListenAndServe()
}

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		pr, pw := io.Pipe()
		tee := io.TeeReader(r.Body, pw)
		r.Body = pr
		go func() {
			body, _ := io.ReadAll(tee)
			defer pw.Close()
			logger.Log.Info("This is the logged request:",
				zap.String("body", string(body)))
		}()

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("resp_status", responseData.status),
			zap.Int("resp_size", responseData.size))
	}
	return http.HandlerFunc(logFn)
}

func (h *HandlersServer) newRouter() http.Handler {
	mux := chi.NewRouter()

	mux.Use(WithLogging)
	mux.Use(middleware.Recoverer)

	mux.Post("/update/", h.mainPageJson)
	mux.Post("/update/{type}/{name}/{value}", h.mainPostPagePlain)
	mux.Post("/update/{type}/", h.mainPageFoundErrors)
	mux.Post("/*", h.mainPageError)
	mux.Get("/value/{type}/{name}", h.mainPageGetPlain)
	mux.Get("/value/", h.mainPageGetJson)
	mux.Get("/", h.mainPage)

	return mux
}

func (h *HandlersServer) mainPageJson(res http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	var jsonstr models.Metrics
	if err := jsonstr.JsonDecode(req.Body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	val, err := h.store.SetValueModel(jsonstr)

	if err != nil {
		if errors.Is(err, service.ErrBadName) {
			http.Error(res, "Invalid request!", http.StatusNotFound)
			return
		}
		http.Error(res, "Invalid request!", http.StatusBadRequest)

		return
	}

	buf, err := val.JsonEncode()
	if err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, buf.String())
}

func (h *HandlersServer) mainPageGetJson(res http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	var jsonstr models.Metrics
	if err := jsonstr.JsonDecode(req.Body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	val, err := h.store.GetValueModel(jsonstr)

	if err != nil {
		if errors.Is(err, service.ErrBadName) || errors.Is(err, repository.ErrNotFoundName) {
			http.Error(res, "Invalid request!", http.StatusNotFound)
			return
		}
		http.Error(res, "Invalid request!", http.StatusBadRequest)

		return
	}

	buf, err := val.JsonEncode()
	if err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, buf.String())

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

	err := h.store.RecievePlainValue(typeVal, name, value)

	if err != nil {
		http.Error(res, "Bad value!", http.StatusBadRequest)
		return
	}

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
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

	val, err := h.store.GetValuePlain(typeVal, name)

	if err != nil {
		http.Error(res, "Not found value!", http.StatusNotFound)
		return
	}

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, val)

}

func (h *HandlersServer) mainPage(res http.ResponseWriter, req *http.Request) {

	if req.URL.String() == "" || req.URL.String() == "/" {

		val, err := h.store.GetAllStore()

		if err != nil {
			http.Error(res, "Not found value!", http.StatusNotFound)
			return
		}

		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		io.WriteString(res, val)

	} else {
		http.Error(res, "Bad request", http.StatusBadRequest)

	}

}
