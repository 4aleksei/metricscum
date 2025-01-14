package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/repository/longtermfile"
	"github.com/4aleksei/metricscum/internal/common/streams/compressors/zipdata"
	"github.com/4aleksei/metricscum/internal/common/streams/encoders/jsonencdec"
	"github.com/4aleksei/metricscum/internal/common/streams/sources/singlefile"
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

	l, err := logger.NewLog(cfg.Level)
	if err != nil {
		return err
	}
	fileWork := longtermfile.NewLongTerm(singlefile.NewReader(cfg.FilePath),
		jsonencdec.NewReader(), singlefile.NewWriter(cfg.FilePath), jsonencdec.NewWriter())

	fileWork.UseForWriter(zipdata.NewWriter())
	fileWork.UseForReader(zipdata.NewReader())

	store := repository.NewStoreMuxFiles(&cfg.Repcfg, l, fileWork)
	store.DataRun()

	metricsService := service.NewHandlerStore(store)
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
		store.DataWrite()

		err = l.Sync()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Logger has been flushed")
		}

		os.Exit(1)
	}()

	return server.Serve()
}
