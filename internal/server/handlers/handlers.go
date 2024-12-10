package handlers

import (
	"io"
	"net/http"

	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
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

	mux.Post("/update/gauge/{name}/{value}", h.mainPageGauge)
	mux.Post("/update/counter/{name}/{value}", h.mainPageCounter)
	mux.Post("/update/gauge/", h.mainPageNotFound)
	mux.Post("/update/counter/", h.mainPageNotFound)
	mux.Post("/*", h.mainPageError)

	mux.Get("/value/gauge/{name}", h.mainPageGetGauge)
	mux.Get("/value/counter/{name}", h.mainPageGetCounter)
	mux.Get("/value/*", h.mainPageError)
	mux.Get("/", h.mainPage)

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

		io.WriteString(res, val)

	} else {
		http.Error(res, "Bad request", http.StatusBadRequest)

	}

}
