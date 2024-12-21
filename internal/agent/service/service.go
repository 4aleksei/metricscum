package service

import (
	"errors"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

type AgentMetricsStorage interface {
	Add(string, valuemetric.ValueMetric) valuemetric.ValueMetric
	ReadAllClearCounters(memstorage.FuncReadAllMetric) error
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
	valMetric := valuemetric.ConvertToFloatValueMetric(val)
	_ = h.store.Add(name, *valMetric)
}

func (h *HandlerStore) SetCounter(name string, val int64) {
	valMetric := valuemetric.ConvertToIntValueMetric(val)
	_ = h.store.Add(name, *valMetric)
}

func (h *HandlerStore) RangeMetrics(prog func(string) error) error {
	err := h.store.ReadAllClearCounters(func(key string, val valuemetric.ValueMetric) error {
		typename, valstr := valuemetric.ConvertValueMetricToPlain(val)
		data := typename + "/" + key + "/" + valstr
		return prog(data)
	})
	return err
}

func (h *HandlerStore) RangeMetricsJSON(prog func(*models.Metrics) error) error {
	err := h.store.ReadAllClearCounters(func(key string, val valuemetric.ValueMetric) error {
		valNewModel := new(models.Metrics)
		valNewModel.ConvertMetricToModel(key, val)
		return prog(valNewModel)
	})
	return err
}
