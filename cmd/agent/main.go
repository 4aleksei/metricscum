package main

import (
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/gather"
	"github.com/4aleksei/metricscum/internal/agent/handlers"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/repository/memstoragemux"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	setupFX().Run()
}

func registerSetLoggerLevel(ll *logger.Logger, cfg *config.Config, lc fx.Lifecycle) {
	ll.SetLevel(cfg.Lcfg.Level)
	lc.Append(utils.ToHook(ll))
}

func registerRunnersGather(gg *gather.AppGather, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(gg))
}

func registerRunnersHTTPClient(cc *handlers.App, lc fx.Lifecycle) {
	lc.Append(utils.ToHook(cc))
}

func setupFX() *fx.App {
	app := fx.New(
		fx.Supply(logger.Config{Level: "debug"}),
		fx.StopTimeout(1*time.Minute),
		fx.Provide(
			logger.NewLogger,
			config.GetConfig,
			fx.Annotate(memstoragemux.NewStoreMux,
				fx.As(new(service.AgentMetricsStorage))),
			service.NewHandlerStore,
			gather.NewAppGather,
			handlers.NewApp,
		),

		fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.L}
		}),
		fx.Invoke(
			registerSetLoggerLevel,
			registerRunnersHTTPClient,
			registerRunnersGather,
		),
	)
	return app
}
