package service

import (
	"errors"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository"
)

type AgentMetricsStorage interface {
	Add(string, repository.ValueMetric) repository.ValueMetric
	ReadAllClearCounters(repository.FuncReadAllMetric) error
}

type HandlerStore struct {
	store AgentMetricsStorage
}

func NewHandlerStore(store AgentMetricsStorage) *HandlerStore {
	p := new(HandlerStore)
	p.store = store
	return p
}

var (
	ErrBadValue = errors.New("invalid value")
)

func (h *HandlerStore) SetGauge(name string, val float64) {
	valMetric := repository.ConvertToFloatValueMetric(val)
	_ = h.store.Add(name, *valMetric)
}

func (h *HandlerStore) SetCounter(name string, val int64) {
	valMetric := repository.ConvertToIntValueMetric(val)
	_ = h.store.Add(name, *valMetric)
}

func (h *HandlerStore) RangeMetrics(prog func(string) error) error {
	err := h.store.ReadAllClearCounters(func(key string, val repository.ValueMetric) error {
		typename, valstr := repository.ConvertValueMetricToPlain(val)
		data := typename + "/" + key + "/" + valstr
		return prog(data)
	})
	return err
}

func (h *HandlerStore) RangeMetricsJSON(prog func(*models.Metrics) error) error {
	err := h.store.ReadAllClearCounters(func(key string, val repository.ValueMetric) error {
		valNewModel := new(models.Metrics)
		valNewModel.ConvertMetricToModel(key, val)
		return prog(valNewModel)
	})
	return err
}
