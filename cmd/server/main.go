package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/handlers"
	"github.com/4aleksei/metricscum/internal/server/resources"
	"github.com/4aleksei/metricscum/internal/server/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()
	l, err := logger.NewLog(cfg.Level)
	if err != nil {
		return err
	}

	storageRes, err := resources.CreateResouces(cfg, l)
	if err != nil {
		return err
	}

	metricsService := service.NewHandlerStore(storageRes.Store)
	server := handlers.NewHandlers(metricsService, cfg, l)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println()
		log.Println("Server is shutting down...", sig)

		err := server.Srv.Shutdown(context.TODO())
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Server has been stopped")
		}

		err = storageRes.Close()

		if err != nil {
			log.Println(err)
		}

		err = l.Sync()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Logger has been flushed")
		}
		os.Exit(0)
	}()
	return server.Serve()
}
