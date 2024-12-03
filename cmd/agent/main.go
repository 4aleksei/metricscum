package main

import (
	"log"

	"github.com/4aleksei/metricscum/internal/agent/gather"
	"github.com/4aleksei/metricscum/internal/agent/handlers"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/repository"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	//	cfg := config.GetConfig()

	store := repository.NewStoreMux()
	metricsService := service.NewHandlerStore(store)
	go gather.MainGather(metricsService)
	return handlers.MainHttpClient(metricsService)
}
