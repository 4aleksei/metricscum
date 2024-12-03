package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/stretchr/testify/assert"

	// "github.com/stretchr/testify/require"
	"github.com/4aleksei/metricscum/internal/common/repository"
)

func Test_handlers_mainPageCounter(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
	}
	type request struct {
		method string
		url    string
	}
	store := service.NewHandlerStore(repository.NewStore()) //

	tests := []struct {
		name string
		req  request
		want want
	}{
		{name: "Test No1", req: request{method: http.MethodPost, url: "/update/gauge/test1/10"}, want: want{statusCode: http.StatusOK, contentType: "text/plain; charset=utf-8"}},

		{name: "Test No2", req: request{method: http.MethodPost, url: "/update/counter/test2/10"}, want: want{statusCode: http.StatusOK, contentType: "text/plain; charset=utf-8"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				store: store,
			}
			request := httptest.NewRequest(tt.req.method, tt.req.url, nil)
			w := httptest.NewRecorder()
			h.mainPageCounter(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

		})
	}
}
