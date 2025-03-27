package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // подключаем пакет pprof
	"time"
)

const (
	addr = ":8086" // адрес сервера
)

func startHTTProfile() {
	go func() {
		srv := &http.Server{
			Addr:              addr,
			ReadHeaderTimeout: 2 * time.Second,
		}
		err := srv.ListenAndServe() // запускаем сервер
		if err != nil {
			fmt.Println(err)
		}
	}()
}
