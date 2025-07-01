// Package httpprof
package httpprof

import (
	"context"
	"net/http"
	_ "net/http/pprof" //nolint:gosec  //подключаем пакет pprof
	"sync"
	"time"
)

const (
	addrDefault = ":8085" // адрес сервера
)

type (
	HTTPprof struct {
		httpServerExitDone *sync.WaitGroup
		srv                *http.Server
	}
)

func NewHTTPprof() *HTTPprof {
	return &HTTPprof{
		httpServerExitDone: &sync.WaitGroup{},
		srv: &http.Server{
			Addr:              addrDefault,
			ReadHeaderTimeout: 2 * time.Second},
	}
}

func (h *HTTPprof) Start(ctx context.Context) error {
	go func() {
		defer h.httpServerExitDone.Done()
		h.httpServerExitDone.Add(1)
		if err := h.srv.ListenAndServe(); err != http.ErrServerClosed {
			return
		}
	}()
	return nil
}

func (h *HTTPprof) Stop(ctx context.Context) error {
	if err := h.srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	h.httpServerExitDone.Wait()
	return nil
}
