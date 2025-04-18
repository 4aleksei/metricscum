package gatherps

import (
	"context"

	"strconv"
	"sync"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type AppGatherMem struct {
	serv   *service.HandlerStore
	l      *logger.Logger
	cfg    *config.Config
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func NewGather(serv *service.HandlerStore, l *logger.Logger, cfg *config.Config) *AppGatherMem {
	return &AppGatherMem{
		l:    l,
		serv: serv,
		wg:   sync.WaitGroup{},
		cfg:  cfg,
	}
}

func (app *AppGatherMem) Start(ctx context.Context) error {
	ctxCancel, cancel := context.WithCancel(context.Background())
	app.cancel = cancel
	app.wg.Add(1)
	go app.mainGather(ctxCancel)
	return nil
}

func (app *AppGatherMem) Stop(ctx context.Context) error {
	app.cancel()
	app.wg.Wait()
	return nil
}

func (app *AppGatherMem) mainGather(ctx context.Context) {
	defer app.wg.Done()

	app.l.L.Info("Start gatheringPS stats.")

	for {
		utils.SleepCancellable(ctx, time.Duration(app.cfg.PollInterval)*time.Second)
		select {
		case <-ctx.Done():
			return
		default:

			v, _ := mem.VirtualMemory()

			_, _ = app.serv.SetGauge(ctx, "TotalMemory", float64(v.Total))
			_, _ = app.serv.SetGauge(ctx, "FreeMemory", float64(v.Free))
			c, _ := cpu.Percent(0, true)
			var name []string
			for i := range c {
				name = append(name, "CPUutilization"+strconv.Itoa(i+1))
			}
			_, _ = app.serv.SetGaugeMulti(ctx, name, c)
		}
	}
}
