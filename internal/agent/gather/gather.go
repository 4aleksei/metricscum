package gather

import (
	"math/rand"
	"runtime"

	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/service"
)

func MainGather(store *service.HandlerStore, cfg *config.Config) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var m runtime.MemStats
	for {

		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)

		runtime.ReadMemStats(&m)

		service.SetGauge(store.Store, "Alloc", float64(m.Alloc))
		service.SetGauge(store.Store, "BuckHashSys", float64(m.BuckHashSys))
		service.SetGauge(store.Store, "Frees", float64(m.Frees))
		service.SetGauge(store.Store, "GCCPUFraction", float64(m.GCCPUFraction))
		service.SetGauge(store.Store, "GCSys", float64(m.GCSys))
		service.SetGauge(store.Store, "HeapAlloc", float64(m.HeapAlloc))
		service.SetGauge(store.Store, "HeapIdle", float64(m.HeapIdle))
		service.SetGauge(store.Store, "HeapInuse", float64(m.HeapInuse))
		service.SetGauge(store.Store, "HeapObjects", float64(m.HeapObjects))
		service.SetGauge(store.Store, "HeapReleased", float64(m.HeapReleased))
		service.SetGauge(store.Store, "HeapSys", float64(m.HeapSys))
		service.SetGauge(store.Store, "LastGC", float64(m.LastGC))
		service.SetGauge(store.Store, "Lookups", float64(m.Lookups))
		service.SetGauge(store.Store, "MCacheInuse", float64(m.MCacheInuse))
		service.SetGauge(store.Store, "MCacheSys", float64(m.MCacheSys))
		service.SetGauge(store.Store, "MSpanInuse", float64(m.MSpanInuse))
		service.SetGauge(store.Store, "MSpanSys", float64(m.MSpanSys))
		service.SetGauge(store.Store, "Mallocs", float64(m.Mallocs))
		service.SetGauge(store.Store, "NextGC", float64(m.NextGC))
		service.SetGauge(store.Store, "NumForcedGC", float64(m.NumForcedGC))
		service.SetGauge(store.Store, "NumGC", float64(m.NumGC))
		service.SetGauge(store.Store, "OtherSys", float64(m.OtherSys))
		service.SetGauge(store.Store, "PauseTotalNs", float64(m.PauseTotalNs))
		service.SetGauge(store.Store, "StackInuse", float64(m.StackInuse))
		service.SetGauge(store.Store, "StackSys", float64(m.StackSys))
		service.SetGauge(store.Store, "Sys", float64(m.Sys))
		service.SetGauge(store.Store, "TotalAlloc", float64(m.TotalAlloc))

		service.SetCounter(store.Store, "PollCount", 1)
		service.SetGaugeFloat(store.Store, "RandomValue", r.Float64())

	}

}
