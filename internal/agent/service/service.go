package service

import (
	"context"
	"errors"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

type AgentMetricsStorage interface {
	Add(context.Context, string, valuemetric.ValueMetric) (valuemetric.ValueMetric, error)
	ReadAllClearCounters(context.Context, memstorage.FuncReadAllMetric) error
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

func (h *HandlerStore) SetGauge(ctx context.Context, name string, val float64) {
	valMetric := valuemetric.ConvertToFloatValueMetric(val)
	_, _ = h.store.Add(ctx, name, *valMetric)
}

func (h *HandlerStore) SetCounter(ctx context.Context, name string, val int64) {
	valMetric := valuemetric.ConvertToIntValueMetric(val)
	_, _ = h.store.Add(ctx, name, *valMetric)
}

func (h *HandlerStore) RangeMetrics(ctx context.Context, prog func(context.Context, string) error) error {
	err := h.store.ReadAllClearCounters(ctx, func(key string, val valuemetric.ValueMetric) error {
		typename, valstr := valuemetric.ConvertValueMetricToPlain(val)
		data := typename + "/" + key + "/" + valstr
		return prog(ctx, data)
	})
	return err
}

func (h *HandlerStore) RangeMetricsJSON(ctx context.Context, prog func(context.Context, *models.Metrics) error) error {
	err := h.store.ReadAllClearCounters(ctx, func(key string, val valuemetric.ValueMetric) error {
		valNewModel := new(models.Metrics)
		valNewModel.ConvertMetricToModel(key, val)
		return prog(ctx, valNewModel)
	})
	return err
}

func (h *HandlerStore) RangeMetricsJSONS(ctx context.Context, prog func(context.Context, *[]models.Metrics) error) error {
	resmodels := new([]models.Metrics)
	err := h.store.ReadAllClearCounters(ctx, func(key string, val valuemetric.ValueMetric) error {
		var valNewModel models.Metrics
		valNewModel.ConvertMetricToModel(key, val)
		*resmodels = append(*resmodels, valNewModel)
		return nil
	})
	if err != nil {
		return err
	}
	return prog(ctx, resmodels)
}
