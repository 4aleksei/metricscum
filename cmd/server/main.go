package main

import (
	"log"

	"os"
	"os/signal"
	"syscall"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/datawriter"
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
	if err := logger.Initialize(cfg.Level); err != nil {
		return err
	}

	store := repository.NewStoreMux()
	dWriter := datawriter.NewAppWriter(store, cfg)

	if cfg.Restore {
		dWriter.ReadData()
	}

	dWriter.RunRutine()

	metricsService := service.NewHandlerStore(store)
	server := handlers.NewHandlers(metricsService, cfg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println()
		log.Println("Server is shutting down...", sig)
		dWriter.DoWriteData()
		os.Exit(1)
	}()

	return server.Serve()
}
