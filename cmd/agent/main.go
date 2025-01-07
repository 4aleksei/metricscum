package main

import (
	"log"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/gather"
	"github.com/4aleksei/metricscum/internal/agent/handlers"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/repository/memstoragemux"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()
	store := memstoragemux.NewStoreMux()
	metricsService := service.NewHandlerStore(store)

	gatherApp := gather.NewAppGather(metricsService, cfg)
	gatherApp.RunRutine()
	mainHTTPClient := handlers.NewApp(metricsService, cfg)
	return mainHTTPClient.Run()
}
