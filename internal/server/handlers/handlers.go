package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/4aleksei/metricscum/internal/common/logger"
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

func (h *HandlersServer) newRouter() http.Handler {
	mux := chi.NewRouter()

	mux.Use(httplogs.WithLogging)
	mux.Use(httpgzip.GzipMiddleware)
	mux.Use(middleware.Recoverer)

	mux.Post("/update/", h.mainPageJSON)
	mux.Post("/update/{type}/{name}/{value}", h.mainPostPagePlain)
	mux.Post("/update/{type}/", h.mainPageFoundErrors)
	mux.Post("/*", h.mainPageError)
	mux.Get("/value/{type}/{name}", h.mainPageGetPlain)
	mux.Post("/value/", h.mainPageGetJSON)
	mux.Get("/", h.mainPage)

	return mux
}

func (h *HandlersServer) mainPageJSON(res http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	var JSONstr models.Metrics
	if err := JSONstr.JSONDecode(req.Body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	val, err := h.store.SetValueModel(JSONstr)

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
		logger.Log.Debug("error encoding response", zap.Error(errson))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(res, buf.String()); err != nil {
		logger.Log.Debug("error writing response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (h *HandlersServer) mainPageGetJSON(res http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	var JSONstr models.Metrics
	if err := JSONstr.JSONDecode(req.Body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	val, err := h.store.GetValueModel(JSONstr)

	if err != nil {
		if errors.Is(err, service.ErrBadName) || errors.Is(err, memstorage.ErrNotFoundName) {
			http.Error(res, "Invalid request!", http.StatusNotFound)
			return
		}
		http.Error(res, "Invalid request!", http.StatusBadRequest)

		return
	}

	var buf bytes.Buffer
	errson := val.JSONEncodeBytes(io.Writer(&buf))

	if errson != nil {
		logger.Log.Debug("error encoding response", zap.Error(errson))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(res, buf.String()); err != nil {
		logger.Log.Debug("error writing response", zap.Error(err))
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

	err := h.store.RecievePlainValue(typeVal, name, value)

	if err != nil {
		http.Error(res, "Bad value!", http.StatusBadRequest)
		return
	}

	switch req.Header.Get("Accept") {
	case "text/html":
		res.Header().Add("Content-Type", "text/html")

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

	val, err := h.store.GetValuePlain(name, typeVal)

	if err != nil {
		http.Error(res, "Not found value!", http.StatusNotFound)
		return
	}

	switch req.Header.Get("Accept") {
	case "text/html":
		res.Header().Add("Content-Type", "text/html")

	default:
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")

	}
	res.WriteHeader(http.StatusOK)

	if _, err := io.WriteString(res, val); err != nil {
		logger.Log.Debug("error writing response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (h *HandlersServer) mainPage(res http.ResponseWriter, req *http.Request) {
	if req.URL.String() == "" || req.URL.String() == "/" {
		val, err := h.store.GetAllStore()
		if err != nil {
			http.Error(res, "Not found value!", http.StatusNotFound)
			return
		}

		switch req.Header.Get("Accept") {
		case "text/html":
			res.Header().Add("Content-Type", "text/html")

		default:
			res.Header().Add("Content-Type", "text/plain; charset=utf-8")

		}
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write([]byte(val)); err != nil {
			logger.Log.Debug("error writing response", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(res, "Bad request", http.StatusBadRequest)
	}
}
