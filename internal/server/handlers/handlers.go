package handlers

import (
	"net/http"
	"strings"

	"github.com/4aleksei/metricscum/internal/server/service"
)

type handlers struct {
	store *service.HandlerStore
}

func newHandlers(store *service.HandlerStore) *handlers {
	return &handlers{
		store: store,
	}
}

func Serve(store *service.HandlerStore) error {
	h := newHandlers(store)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    `:8080`,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *handlers) *http.ServeMux {
	mux := http.NewServeMux()
	//	mux.HandleFunc("POST /update/{t}/{n}/{v}", h.Update)
	//	mux.HandleFunc("GET /value/{t}/{n}", h.Get)
	//	mux.HandleFunc("GET /", h.TypeNames)

	mux.HandleFunc(`/update/gauge/`, h.mainPageGauge)
	mux.HandleFunc(`/update/counter/`, h.mainPageCounter)
	mux.HandleFunc(`/update/`, h.mainPageError)

	return mux
}

func (h *handlers) mainPageGauge(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	urlPart := strings.Split(req.URL.Path, "/")

	if len(urlPart) == 4 && urlPart[3] == "" {
		http.Error(res, "Bad data!", http.StatusNotFound)
		return
	}

	if len(urlPart) != 5 {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	err := service.RecieveGauge(h.store.Store, urlPart[3], urlPart[4])

	if err != nil {
		http.Error(res, "Bad gauge value!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")

}

func (h *handlers) mainPageCounter(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	urlPart := strings.Split(req.URL.Path, "/")

	if len(urlPart) == 4 && urlPart[3] == "" {
		http.Error(res, "Bad data!", http.StatusNotFound)
		return
	}

	if len(urlPart) != 5 {
		http.Error(res, "Bad type!", http.StatusBadRequest)
		return
	}

	err := service.RecieveCounter(h.store.Store, urlPart[3], urlPart[4])

	if err != nil {
		http.Error(res, "Bad counter value!", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
}

func (h *handlers) mainPageError(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	http.Error(res, "Bad request", http.StatusBadRequest)
}
