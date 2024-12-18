package main

import (
	"log"

	"github.com/4aleksei/metricscum/internal/common/logger"
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
	logger.Initialize(cfg.Level)
	store := repository.NewStoreMux()
	metricsService := service.NewHandlerStore(store)

	server := handlers.NewHandlers(metricsService, cfg)
	return server.Serve()
}
