package gather

import (
	"math/rand"
	"runtime"

	"time"

	"github.com/4aleksei/metricscum/internal/agent/service"
)

func MainGather(store *service.HandlerStore) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var m runtime.MemStats
	for {

		time.Sleep(2 * time.Second)

		runtime.ReadMemStats(&m)

		service.SetGauge(store.Store, "Alloc", m.Alloc)
		service.SetGauge(store.Store, "BuckHashSys", m.BuckHashSys)
		service.SetGauge(store.Store, "Frees", m.Frees)
		service.SetGauge(store.Store, "GCCPUFraction", uint64(m.GCCPUFraction))
		service.SetGauge(store.Store, "GCSys", m.GCSys)
		service.SetGauge(store.Store, "HeapAlloc", m.HeapAlloc)
		service.SetGauge(store.Store, "HeapIdle", m.HeapIdle)
		service.SetGauge(store.Store, "HeapInuse", m.HeapInuse)
		service.SetGauge(store.Store, "HeapObjects", m.HeapObjects)
		service.SetGauge(store.Store, "HeapReleased", m.HeapReleased)
		service.SetGauge(store.Store, "HeapSys", m.HeapSys)
		service.SetGauge(store.Store, "LastGC", m.LastGC)
		service.SetGauge(store.Store, "Lookups", m.Lookups)
		service.SetGauge(store.Store, "MCacheInuse", m.MCacheInuse)
		service.SetGauge(store.Store, "MCacheSys", m.MCacheSys)
		service.SetGauge(store.Store, "MSpanInuse", m.MSpanInuse)
		service.SetGauge(store.Store, "MSpanSys", m.MSpanSys)
		service.SetGauge(store.Store, "Mallocs", m.Mallocs)
		service.SetGauge(store.Store, "NextGC", m.NextGC)
		service.SetGauge(store.Store, "NumForcedGC", uint64(m.NumForcedGC))
		service.SetGauge(store.Store, "NumGC", uint64(m.NumGC))
		service.SetGauge(store.Store, "OtherSys", m.OtherSys)
		service.SetGauge(store.Store, "PauseTotalNs", m.PauseTotalNs)
		service.SetGauge(store.Store, "StackInuse", m.StackInuse)
		service.SetGauge(store.Store, "StackSys", m.StackSys)
		service.SetGauge(store.Store, "Sys", m.Sys)
		service.SetGauge(store.Store, "TotalAlloc", m.TotalAlloc)

		service.SetCounter(store.Store, "PollCount", 1)
		service.SetGaugeFloat(store.Store, "RandomValue", r.Float64())

	}

}
