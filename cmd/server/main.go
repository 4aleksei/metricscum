package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/4aleksei/metricscum/cmd/server/migrate"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/handlers"
	"github.com/4aleksei/metricscum/internal/server/resources"
	"github.com/4aleksei/metricscum/internal/server/service"
	"go.uber.org/zap"
)

const (
	defaultHTTPshutdown int = 10
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

	if cfg.DBcfg.DatabaseDSN != "" {
		err := migrate.Migrate(l, cfg.DBcfg.DatabaseDSN, "up")
		if err != nil {
			l.Error("Error goose UP migration:", zap.Error(err))
			return err
		}
	}

	storageRes, err := resources.CreateResouces(cfg, l)
	if err != nil {
		l.Error("Error cretae resources :", zap.Error(err))
		return err
	}

	metricsService := service.NewHandlerStore(storageRes.Store)
	server := handlers.NewHandlers(metricsService, cfg, l)

	server.Serve()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs

	l.Info("Server is shutting down...", zap.String("signal", sig.String()))

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), time.Duration(defaultHTTPshutdown)*time.Second)
	defer shutdownRelease()

	if err := server.Srv.Shutdown(shutdownCtx); err != nil {
		l.Error("HTTP shutdown error :", zap.Error(err))
	} else {
		l.Info("Server shutdown complete")
	}

	err = storageRes.Close(context.Background())
	if err != nil {
		l.Error("Resources close error :", zap.Error(err))
	} else {
		l.Info("Resources close complete")
	}

	_ = l.Sync()

	return nil
}
