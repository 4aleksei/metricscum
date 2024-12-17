package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/stretchr/testify/assert"

	"io"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body string, contentType string) (*http.Response, string) {
	var buf bytes.Buffer
	buf.WriteString(body)
	req, err := http.NewRequest(method, ts.URL+path, &buf)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_handlers_mainPagePlain(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	type request struct {
		method string
		url    string
	}
	store := service.NewHandlerStore(repository.NewStore())
	h := new(HandlersServer)
	h.store = store

	ts := httptest.NewServer(h.newRouter())
	defer ts.Close()

	tests := []struct {
		name string
		req  request
		want want
	}{
		{name: "Test No1", req: request{method: http.MethodPost, url: "/update/gauge/test1/10"}, want: want{statusCode: http.StatusOK, contentType: "text/plain; charset=utf-8"}},

		{name: "Test No2", req: request{method: http.MethodPost, url: "/update/counter/test2/10"}, want: want{statusCode: http.StatusOK, contentType: "text/plain; charset=utf-8"}},

		{name: "Test No3", req: request{method: http.MethodPost, url: "/update/counter/"}, want: want{statusCode: http.StatusNotFound, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No4", req: request{method: http.MethodPost, url: "/update/gauge/"}, want: want{statusCode: http.StatusNotFound, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No5", req: request{method: http.MethodPost, url: "/update/unknown/"}, want: want{statusCode: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No6", req: request{method: http.MethodPost, url: "/update/unknown/test3/10"}, want: want{statusCode: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No7", req: request{method: http.MethodPost, url: "/update/counter//10"}, want: want{statusCode: http.StatusNotFound, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No8", req: request{method: http.MethodPost, url: "/update/gauge/test3/dfdfs"}, want: want{statusCode: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No9", req: request{method: http.MethodPost, url: "/update/counter/test4/5454.3434"}, want: want{statusCode: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
		{name: "Test No10", req: request{method: http.MethodPost, url: "/update/counter/testreal/10"}, want: want{statusCode: http.StatusOK, contentType: "text/plain; charset=utf-8"}},

		{name: "Test No11", req: request{method: http.MethodGet, url: "/value/counter/testreal"}, want: want{statusCode: http.StatusOK, contentType: "text/plain; charset=utf-8", body: "10"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, respBody := testRequest(t, ts, tt.req.method, tt.req.url, "", "")

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			}

			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, respBody)
			}

			resp.Body.Close()
		})
	}
}

func Test_handlers_mainPageJSON(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	type request struct {
		method      string
		url         string
		body        string
		contentType string
	}
	store := service.NewHandlerStore(repository.NewStore())
	h := new(HandlersServer)
	h.store = store

	ts := httptest.NewServer(h.newRouter())
	defer ts.Close()

	/*
	   type Metrics struct {
	      ID    string   `JSON:"id"`              // имя метрики
	      MType string   `JSON:"type"`            // параметр, принимающий значение gauge или counter
	      Delta *int64   `JSON:"delta,omitempty"` // значение метрики в случае передачи counter
	      Value *float64 `JSON:"value,omitempty"` // значение метрики в случае передачи gauge
	   }
	*/

	tests := []struct {
		name string
		req  request
		want want
	}{
		{name: "JSON Test No1", req: request{method: http.MethodPost, url: "/update/", body: " {\"id\":\"test1\" , \"type\":\"counter\" , \"delta\": 100 }  ", contentType: "application/json"}, want: want{statusCode: http.StatusOK, contentType: "application/json", body: " {\"id\":\"test1\" , \"type\":\"counter\" , \"delta\":100 }  "}},
		{name: "JSON Test No2", req: request{method: http.MethodPost, url: "/update/", body: " {\"id\":\"\" , \"type\":\"counter\" , \"delta\": 100 }  ", contentType: "application/json"}, want: want{statusCode: http.StatusNotFound, contentType: "", body: ""}},
		{name: "JSON Test No3", req: request{method: http.MethodPost, url: "/value/", body: " {\"id\":\"test1\" , \"type\":\"counter\" }  ", contentType: "application/json"}, want: want{statusCode: http.StatusOK, contentType: "application/json", body: " {\"id\":\"test1\" , \"type\":\"counter\" , \"delta\":100 }  "}},
		{name: "JSON Test No4", req: request{method: http.MethodPost, url: "/value/", body: " {\"id\":\"test2\" , \"type\":\"counter\" }  ", contentType: "application/json"}, want: want{statusCode: http.StatusNotFound, contentType: "", body: ""}},

		//{name: "Test No3", req: request{method: http.MethodPost, url: "/update/"}, want: want{statusCode: http.StatusNotFound, contentType: "text/plain; charset=utf-8"}},
		//{name: "Test No4", req: request{method: http.MethodPost, url: "/update/"}, want: want{statusCode: http.StatusNotFound, contentType: "text/plain; charset=utf-8"}},
		//{name: "Test No5", req: request{method: http.MethodPost, url: "/update/"}, want: want{statusCode: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, respBody := testRequest(t, ts, tt.req.method, tt.req.url, tt.req.body, tt.req.contentType)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			}

			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, respBody)
			}
			resp.Body.Close()
		})
	}
}
