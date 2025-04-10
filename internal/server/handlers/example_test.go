package handlers

import (
	"fmt"
	"io"
	"net/http/httptest"

	"context"

	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/go-chi/chi/v5"
)

func Example_mainPage() { //endpoint GET "/"
	h := new(HandlersServer)
	req := httptest.NewRequest("GET", "/bad", nil)
	w := httptest.NewRecorder()
	h.mainPage(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// Output:
	// 400
	// text/plain; charset=utf-8
	// Bad request
}

func Example_mainPageJSON() { //endpoint POST "/update/"
	h := new(HandlersServer)
	req := httptest.NewRequest("POST", "/update/", nil)
	w := httptest.NewRecorder()
	h.mainPageJSON(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// Output:
	// 400
	// text/plain; charset=utf-8
	// Bad type!
}

func Example_mainPostPagePlain() { //endpoint POST "/update/{type}/{name}/{value}"
	store := service.NewHandlerStore(memstorage.NewStore())
	h := new(HandlersServer)
	h.store = store

	r := httptest.NewRequest("POST", "/update/counter/test2/100", nil)
	w := httptest.NewRecorder()
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add("type", fmt.Sprintf("%v", "counter"))
	chiCtx.URLParams.Add("name", fmt.Sprintf("%v", "test2"))
	chiCtx.URLParams.Add("value", fmt.Sprintf("%v", "100"))

	h.mainPostPagePlain(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// Output:
	// 200
	// text/plain; charset=utf-8
}
