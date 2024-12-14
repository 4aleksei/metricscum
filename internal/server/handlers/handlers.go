package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
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

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
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

	mux.Post("/update/gauge/{name}/{value}", WithLogging(h.mainPageGauge))
	mux.Post("/update/counter/{name}/{value}", WithLogging(h.mainPageCounter))
	mux.Post("/update/gauge/", WithLogging(h.mainPageNotFound))
	mux.Post("/update/counter/", WithLogging(h.mainPageNotFound))
	mux.Post("/*", WithLogging(h.mainPageError))
	mux.Get("/value/gauge/{name}", WithLogging(h.mainPageGetGauge))
	mux.Get("/value/counter/{name}", WithLogging(h.mainPageGetCounter))
	mux.Get("/value/*", WithLogging(h.mainPageError))
	mux.Get("/", WithLogging(h.mainPage))

	return mux
}

func (h *HandlersServer) mainPageError(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Bad request", http.StatusBadRequest)
}

func (h *HandlersServer) mainPageNotFound(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Not Found", http.StatusNotFound)
}

func (h *HandlersServer) mainPageGauge(res http.ResponseWriter, req *http.Request) {

	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	if value == "" {
		http.Error(res, "Bad data!", http.StatusNotFound)
		return
	}

	if name == "" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	err := h.store.RecieveGauge(name, value)

	if err != nil {
		http.Error(res, "Bad gauge value!", http.StatusBadRequest)
		return
	}

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)

}

func (h *HandlersServer) mainPageCounter(res http.ResponseWriter, req *http.Request) {

	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	if value == "" {
		http.Error(res, "Bad data!", http.StatusNotFound)
		return
	}

	if name == "" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	err := h.store.RecieveCounter(name, value)

	if err != nil {
		http.Error(res, "Bad counter value!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)

}

func (h *HandlersServer) mainPageGetGauge(res http.ResponseWriter, req *http.Request) {

	name := chi.URLParam(req, "name")

	if name == "" {
		http.Error(res, "Bad type!", http.StatusNotFound)
		return
	}

	val, err := h.store.GetGauge(name)

	if err != nil {
		http.Error(res, "Not found value!", http.StatusNotFound)
		return
	}

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, val)

}

func (h *HandlersServer) mainPageGetCounter(res http.ResponseWriter, req *http.Request) {

	name := chi.URLParam(req, "name")

	if name == "" {
		http.Error(res, "Bad type!", http.StatusNotFound)
		return
	}

	val, err := h.store.GetCounter(name)

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
