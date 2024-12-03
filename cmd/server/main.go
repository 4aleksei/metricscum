package main

import (
	"log"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/handlers"
	"github.com/4aleksei/metricscum/internal/server/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	store := repository.NewStore()
	metricsService := service.NewHandlerStore(store)

	return handlers.Serve(metricsService, cfg)
}
