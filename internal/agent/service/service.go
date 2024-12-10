package service

import (
	"errors"

	"github.com/4aleksei/metricscum/internal/common/repository"
)

type AgentMetricsStorage interface {
	Update(string, repository.GaugeMetric)
	Add(string, repository.CounterMetric)
	ReadAllClearCounters(repository.FuncReadAllMetric) error
}

type HandlerStore struct {
	Store AgentMetricsStorage
}

func NewHandlerStore(store AgentMetricsStorage) *HandlerStore {

	p := new(HandlerStore)
	p.Store = store

	return p
}

var (
	ErrBadValue = errors.New("invalid value")
)

func (h *HandlerStore) SetGauge(name string, val float64) {

	h.Store.Update(name, repository.GaugeMetric(val))

}

func (h *HandlerStore) SetGaugeFloat(name string, val float64) {

	h.Store.Update(name, repository.GaugeMetric(val))

}

func (h *HandlerStore) SetCounter(name string, val uint64) {

	h.Store.Add(name, repository.CounterMetric(val))

}

func (h *HandlerStore) RangeMetrics(prog func(string) error) error {

	err := h.Store.ReadAllClearCounters(func(typename string, name string, value string) error {
		data := typename + "/" + name + "/" + value
		return prog(data)
	})

	return err
}
