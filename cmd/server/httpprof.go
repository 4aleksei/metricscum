package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
)

const (
	addr = ":8086" // адрес сервера
)

func startHTTProfile() {
	go func() {
		fmt.Println("Start Server")           // запускаем полезную нагрузку в фоне
		err := http.ListenAndServe(addr, nil) // запускаем сервер
		if err != nil {
			fmt.Println(err)
		}
	}()
}
