package handlers

import (
	"io"
	"net/http"

	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
)

type handlers struct {
	store *service.HandlerStore
}

func newHandlers(store *service.HandlerStore) *handlers {
	return &handlers{
		store: store,
	}
}

func Serve(store *service.HandlerStore, cfg *config.Config) error {
	h := newHandlers(store)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *handlers) http.Handler {
	//mux := http.NewServeMux()
	mux := chi.NewRouter()

	mux.Post("/update/gauge/{name}/{value}", h.mainPageGauge)
	mux.Post("/update/counter/{name}/{value}", h.mainPageCounter)
	mux.Post("/update/", h.mainPageError)

	mux.Get("/value/gauge/{name}", h.mainPageGetGauge)
	mux.Get("/value/counter/{name}", h.mainPageGetCounter)
	mux.Get("/", h.mainPage)

	//mux.HandleFunc("POST /update/{t}/{n}/{v}", h.Update)
	//mux.HandleFunc("GET /value/{t}/{n}", h.Get)
	//mux.HandleFunc("GET /", h.TypeNames)

	//mux.HandleFunc(`/update/gauge/`, h.mainPageGauge)
	//mux.HandleFunc(`/update/counter/`, h.mainPageCounter)
	//mux.HandleFunc(`/update/`, h.mainPageError)

	return mux
}

func (h *handlers) mainPageError(res http.ResponseWriter, req *http.Request) {

	http.Error(res, "Bad request", http.StatusBadRequest)
}

func (h *handlers) mainPageGauge(res http.ResponseWriter, req *http.Request) {

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

	err := service.RecieveGauge(h.store.Store, name, value)

	if err != nil {
		http.Error(res, "Bad gauge value!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")

}

func (h *handlers) mainPageCounter(res http.ResponseWriter, req *http.Request) {

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

	err := service.RecieveCounter(h.store.Store, name, value)

	if err != nil {
		http.Error(res, "Bad counter value!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
}

func (h *handlers) mainPageGetGauge(res http.ResponseWriter, req *http.Request) {

	name := chi.URLParam(req, "name")

	if name == "" {
		http.Error(res, "Bad type!", http.StatusNotFound)
		return
	}

	val, err := service.GetGauge(h.store.Store, name)

	if err != nil {
		http.Error(res, "Not found value!", http.StatusNotFound)
		return
	}

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(res, val)

}

func (h *handlers) mainPageGetCounter(res http.ResponseWriter, req *http.Request) {

	name := chi.URLParam(req, "name")

	if name == "" {
		http.Error(res, "Bad type!", http.StatusNotFound)
		return
	}

	val, err := service.GetCounter(h.store.Store, name)

	if err != nil {
		http.Error(res, "Not found value!", http.StatusNotFound)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")

	io.WriteString(res, val)

}

func (h *handlers) mainPage(res http.ResponseWriter, req *http.Request) {

	if req.URL.String() == "" || req.URL.String() == "/" {

		val, err := service.GetAllStore(h.store.Store)

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
