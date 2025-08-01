package handlers

import (
	"context"

	"sync"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"

	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/logger"

	"github.com/4aleksei/metricscum/internal/common/utils"
)

type App struct {
	serv   *service.HandlerStore
	cfg    *config.Config
	l      *logger.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewApp(store *service.HandlerStore, l *logger.Logger, cfg *config.Config) *App {
	return &App{
		serv: store,
		cfg:  cfg,
		l:    l,
	}
}

func (app *App) Start(ctx context.Context) error {
	ctxCancel, cancel := context.WithCancel(context.Background())
	app.cancel = cancel
	app.wg = sync.WaitGroup{}
	app.wg.Add(1)
	go app.run(ctxCancel)
	return nil
}

func (app *App) Stop(ctx context.Context) error {
	app.cancel()
	app.wg.Wait()
	app.serv.GracefulStop()
	return nil
}

func (app *App) run(ctx context.Context) {
	defer app.wg.Done()
	app.l.L.Info("Start reporting.")
	for {
		utils.SleepCancellable(ctx, time.Duration(app.cfg.ReportInterval)*time.Second)
		select {
		case <-ctx.Done():
			app.l.L.Info("Stop reporting.")
			return
		default:
			_ = app.serv.SendMetrics(ctx)
		}
	}
}
