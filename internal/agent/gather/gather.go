package gather

import (
	"context"
	"crypto/rand"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/utils"
)

type AppGather struct {
	serv   *service.HandlerStore
	l      *logger.Logger
	cfg    *config.Config
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func NewAppGather(serv *service.HandlerStore, l *logger.Logger, cfg *config.Config) *AppGather {
	return &AppGather{
		l:    l,
		serv: serv,
		cfg:  cfg,
	}
}

func (app *AppGather) Start(ctx context.Context) error {
	ctxCancel, cancel := context.WithCancel(context.Background())
	app.cancel = cancel
	app.wg.Add(1)
	go app.mainGather(ctxCancel)
	return nil
}

func (app *AppGather) Stop(ctx context.Context) error {
	app.cancel()
	app.wg.Wait()
	return nil
}

const maxInt int64 = 1 << 53

func Intn() int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(maxInt))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}

func RandFloat64() float64 {
	return float64(Intn()) / float64(maxInt)
}

func (app *AppGather) mainGather(ctx context.Context) {
	defer app.wg.Done()
	var m runtime.MemStats

	app.l.L.Info("Start gathering stats.")

	for {
		utils.SleepCancellable(ctx, time.Duration(app.cfg.PollInterval)*time.Second)
		select {
		case <-ctx.Done():
			app.l.L.Info("Stop gathering.")
			return
		default:
			runtime.ReadMemStats(&m)
			app.serv.SetGauge(ctx, "Alloc", float64(m.Alloc))
			app.serv.SetGauge(ctx, "BuckHashSys", float64(m.BuckHashSys))
			app.serv.SetGauge(ctx, "Frees", float64(m.Frees))
			app.serv.SetGauge(ctx, "GCCPUFraction", float64(m.GCCPUFraction))
			app.serv.SetGauge(ctx, "GCSys", float64(m.GCSys))
			app.serv.SetGauge(ctx, "HeapAlloc", float64(m.HeapAlloc))
			app.serv.SetGauge(ctx, "HeapIdle", float64(m.HeapIdle))
			app.serv.SetGauge(ctx, "HeapInuse", float64(m.HeapInuse))
			app.serv.SetGauge(ctx, "HeapObjects", float64(m.HeapObjects))
			app.serv.SetGauge(ctx, "HeapReleased", float64(m.HeapReleased))
			app.serv.SetGauge(ctx, "HeapSys", float64(m.HeapSys))
			app.serv.SetGauge(ctx, "LastGC", float64(m.LastGC))
			app.serv.SetGauge(ctx, "Lookups", float64(m.Lookups))
			app.serv.SetGauge(ctx, "MCacheInuse", float64(m.MCacheInuse))
			app.serv.SetGauge(ctx, "MCacheSys", float64(m.MCacheSys))
			app.serv.SetGauge(ctx, "MSpanInuse", float64(m.MSpanInuse))
			app.serv.SetGauge(ctx, "MSpanSys", float64(m.MSpanSys))
			app.serv.SetGauge(ctx, "Mallocs", float64(m.Mallocs))
			app.serv.SetGauge(ctx, "NextGC", float64(m.NextGC))
			app.serv.SetGauge(ctx, "NumForcedGC", float64(m.NumForcedGC))
			app.serv.SetGauge(ctx, "NumGC", float64(m.NumGC))
			app.serv.SetGauge(ctx, "OtherSys", float64(m.OtherSys))
			app.serv.SetGauge(ctx, "PauseTotalNs", float64(m.PauseTotalNs))
			app.serv.SetGauge(ctx, "StackInuse", float64(m.StackInuse))
			app.serv.SetGauge(ctx, "StackSys", float64(m.StackSys))
			app.serv.SetGauge(ctx, "Sys", float64(m.Sys))
			app.serv.SetGauge(ctx, "TotalAlloc", float64(m.TotalAlloc))
			app.serv.SetCounter(ctx, "PollCount", 1)
			app.serv.SetGauge(ctx, "RandomValue", RandFloat64())
		}
	}
}
