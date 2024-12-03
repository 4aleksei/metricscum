package service

import (
	"errors"

	"github.com/4aleksei/metricscum/internal/common/repository"
)

type HandlerStore struct {
	Store repository.MetricsStorage
}

func NewHandlerStore(store repository.MetricsStorage) *HandlerStore {
	return &HandlerStore{
		Store: store,
	}
}

var (
	ErrBadValue = errors.New("invalid value")
)

func SetGauge(store repository.MetricsStorage, name string, val float64) {

	repository.AddGauge(store, name, repository.GaugeMetric(val))

}

func SetGaugeFloat(store repository.MetricsStorage, name string, val float64) {

	repository.AddGauge(store, name, repository.GaugeMetric(val))

}

func SetCounter(store repository.MetricsStorage, name string, val uint64) {

	repository.AddCounter(store, name, repository.CounterMetric(val))

}

func RangeMetrics(store repository.MetricsStorage, prog func(string) error) error {

	repository.ReadAllClearCounters(store, func(typename string, name string, value string) error {
		data := typename + "/" + name + "/" + value

		return prog(data)
	})

	return nil
}
