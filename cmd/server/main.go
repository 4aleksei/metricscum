// Metrics Alerting Service
// Application server
// Receive metrics on endpoint and save it at BD
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/4aleksei/metricscum/cmd/server/migrate"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/server/config"
	grpcmetrics "github.com/4aleksei/metricscum/internal/server/grpcservice"
	"github.com/4aleksei/metricscum/internal/server/handlers"
	"github.com/4aleksei/metricscum/internal/server/resources"
	"github.com/4aleksei/metricscum/internal/server/service"
	"go.uber.org/zap"
)

const (
	defaultHTTPshutdown int = 10
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printVersion() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}

func main() {
	printVersion()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}

	l, err := logger.NewLog(cfg.Level)
	if err != nil {
		return err
	}

	if cfg.DBcfg.DatabaseDSN != "" {
		errM := migrate.Migrate(l, cfg.DBcfg.DatabaseDSN, "up")
		if errM != nil {
			l.Error("Error goose UP migration:", zap.Error(errM))
			return errM
		}
	}

	storageRes, errC := resources.CreateResouces(cfg, l)
	if errC != nil {
		l.Error("Error create resources :", zap.Error(errC))
		return errC
	}

	metricsService := service.NewHandlerStore(storageRes.Store)
	server, errS := handlers.NewServer(metricsService, cfg, l)
	if errS != nil {
		l.Error("Error server construct:", zap.Error(errS))
		return errS
	}

	server.Serve()

	grpcServ, errG := grpcmetrics.NewgPRC(metricsService, cfg, l)
	if errG != nil {
		l.Error("Error server grpc construct:", zap.Error(errG))
		return errG
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-sigs

	l.Info("Server is shutting down...", zap.String("signal", sig.String()))

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), time.Duration(defaultHTTPshutdown)*time.Second)
	defer shutdownRelease()

	if errS := server.Srv.Shutdown(shutdownCtx); errS != nil {
		l.Error("HTTP shutdown error :", zap.Error(errS))
	} else {
		l.Info("Server shutdown complete")
	}

	grpcServ.StopServ()

	errClose := storageRes.Close(context.Background())
	if errClose != nil {
		l.Error("Resources close error :", zap.Error(errClose))
	} else {
		l.Info("Resources close complete")
	}

	_ = l.Sync()

	return nil
}
