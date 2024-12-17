package gather

import (
	"math/rand"
	"runtime"

	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/service"
)

type AppGather struct {
	serv *service.HandlerStore
	cfg  *config.Config
}

func NewAppGather(serv *service.HandlerStore, cfg *config.Config) *AppGather {
	app := new(AppGather)
	app.serv = serv
	app.cfg = cfg
	return app
}

func (app *AppGather) RunRutine() {
	go app.mainGather()
}

func (app *AppGather) mainGather() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var m runtime.MemStats
	for {

		time.Sleep(time.Duration(app.cfg.PollInterval) * time.Second)

		runtime.ReadMemStats(&m)

		app.serv.SetGauge("Alloc", float64(m.Alloc))
		app.serv.SetGauge("BuckHashSys", float64(m.BuckHashSys))
		app.serv.SetGauge("Frees", float64(m.Frees))
		app.serv.SetGauge("GCCPUFraction", float64(m.GCCPUFraction))
		app.serv.SetGauge("GCSys", float64(m.GCSys))
		app.serv.SetGauge("HeapAlloc", float64(m.HeapAlloc))
		app.serv.SetGauge("HeapIdle", float64(m.HeapIdle))
		app.serv.SetGauge("HeapInuse", float64(m.HeapInuse))
		app.serv.SetGauge("HeapObjects", float64(m.HeapObjects))
		app.serv.SetGauge("HeapReleased", float64(m.HeapReleased))
		app.serv.SetGauge("HeapSys", float64(m.HeapSys))
		app.serv.SetGauge("LastGC", float64(m.LastGC))
		app.serv.SetGauge("Lookups", float64(m.Lookups))
		app.serv.SetGauge("MCacheInuse", float64(m.MCacheInuse))
		app.serv.SetGauge("MCacheSys", float64(m.MCacheSys))
		app.serv.SetGauge("MSpanInuse", float64(m.MSpanInuse))
		app.serv.SetGauge("MSpanSys", float64(m.MSpanSys))
		app.serv.SetGauge("Mallocs", float64(m.Mallocs))
		app.serv.SetGauge("NextGC", float64(m.NextGC))
		app.serv.SetGauge("NumForcedGC", float64(m.NumForcedGC))
		app.serv.SetGauge("NumGC", float64(m.NumGC))
		app.serv.SetGauge("OtherSys", float64(m.OtherSys))
		app.serv.SetGauge("PauseTotalNs", float64(m.PauseTotalNs))
		app.serv.SetGauge("StackInuse", float64(m.StackInuse))
		app.serv.SetGauge("StackSys", float64(m.StackSys))
		app.serv.SetGauge("Sys", float64(m.Sys))
		app.serv.SetGauge("TotalAlloc", float64(m.TotalAlloc))

		app.serv.SetCounter("PollCount", 1)
		app.serv.SetGauge("RandomValue", r.Float64())

	}

}
