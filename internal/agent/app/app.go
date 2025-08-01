// Package app
package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/gather"
	"github.com/4aleksei/metricscum/internal/agent/gatherps"
	"github.com/4aleksei/metricscum/internal/agent/handlers"
	"github.com/4aleksei/metricscum/internal/agent/poolclients"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/httpprof"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/repository/memstoragemux"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func registerSetLoggerLevel(ll *logger.Logger, cfg *config.Config, lc fx.Lifecycle) {
	ll.SetLevel(cfg.Lcfg.Level)
	lc.Append(utils.ToHook(ll))
}

func registerRunnersGather(gg *gather.AppGather, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(gg))
}

func registerShutdowner(shutdowner fx.Shutdowner, ll *logger.Logger) {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		sig := <-sigs
		ll.L.Info("Receive signal", zap.String("signal", sig.String()))
		if err := shutdowner.Shutdown(); err != nil {
			ll.L.Debug("encountered error trying to shutdown app: ", zap.Error(err))
		}
	}()
}

func registerRunnersGatherps(gg *gatherps.AppGatherMem, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(gg))
}

func registerRunnersSender(cc *handlers.App, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(cc))
}

func registerHTTPprof(hh *httpprof.HTTPprof, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(hh))
}

// SetupFX - application constructor
func SetupFX() *fx.App {
	app := fx.New(
		fx.Supply(logger.Config{Level: "debug"}),
		fx.StopTimeout(1*time.Minute),
		fx.Provide(
			logger.NewLogger,
			config.NewConfig,
			httpprof.NewHTTPprof,
			fx.Annotate(memstoragemux.NewStoreMux,
				fx.As(new(service.AgentMetricsStorage))),
			service.NewHandlerStore,
			gather.NewAppGather,
			gatherps.NewGather,
			poolclients.NewPoolClient,
			//httpclientpool.NewHandler,
			handlers.NewApp,
		),

		fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.L}
		}),
		fx.Invoke(
			registerSetLoggerLevel,
			registerHTTPprof,
			registerRunnersSender,
			registerRunnersGather,
			registerRunnersGatherps,
			registerShutdowner,
		),
	)
	return app
}
