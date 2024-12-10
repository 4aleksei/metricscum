package handlers

import (
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
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_handlers_mainPageCounter(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
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
		{name: "Test No5", req: request{method: http.MethodPost, url: "/update/unknown/test3/10"}, want: want{statusCode: http.StatusBadRequest, contentType: "text/plain; charset=utf-8"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, _ := testRequest(t, ts, tt.req.method, tt.req.url)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			resp.Body.Close()
		})
	}
}
