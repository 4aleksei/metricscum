package service

import (
	"context"
	"errors"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool/job"
	"go.uber.org/zap"
)

type AgentMetricsStorage interface {
	Add(context.Context, string, valuemetric.ValueMetric) (valuemetric.ValueMetric, error)
	ReadAllClearCounters(context.Context, memstorage.FuncReadAllMetric) error
	AddMulti(context.Context, []models.Metrics) ([]models.Metrics, error)
}

type HandlerStore struct {
	store AgentMetricsStorage
	pool  *httpclientpool.PoolHandler
	cfg   *config.Config
	l     *logger.Logger
}

func NewHandlerStore(store AgentMetricsStorage, pool *httpclientpool.PoolHandler, cfg *config.Config, l *logger.Logger) *HandlerStore {
	return &HandlerStore{
		store: store,
		pool:  pool,
		cfg:   cfg,
		l:     l,
	}
}

var (
	ErrBadValue = errors.New("invalid value")
)

func (h *HandlerStore) SetGauge(ctx context.Context, name string, val float64) {
	valMetric := valuemetric.ConvertToFloatValueMetric(val)
	_, _ = h.store.Add(ctx, name, *valMetric)
}

func (h *HandlerStore) SetGaugeMulti(ctx context.Context, name []string, valArr []float64) {
	var valMetric []models.Metrics
	for i, val := range valArr {
		var v models.Metrics
		v.ConvertMetricToModel(name[i], *valuemetric.ConvertToFloatValueMetric(val))
		valMetric = append(valMetric, v)
	}
	_, _ = h.store.AddMulti(ctx, valMetric)
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

func (h *HandlerStore) RangeMetricsJSONS(ctx context.Context, prog func(context.Context, []models.Metrics) error) error {
	resmodels := make([]models.Metrics, 0, 1)
	err := h.store.ReadAllClearCounters(ctx, func(key string, val valuemetric.ValueMetric) error {
		var valNewModel models.Metrics
		valNewModel.ConvertMetricToModel(key, val)
		resmodels = append(resmodels, valNewModel)
		return nil
	})
	if err != nil {
		return err
	}
	return prog(ctx, resmodels)
}

func (h *HandlerStore) SendMetrics(ctx context.Context) error {
	resmodelsTX := make([]models.Metrics, 0, 1)
	err := h.store.ReadAllClearCounters(ctx, func(key string, val valuemetric.ValueMetric) error {
		var valNewModel models.Metrics
		valNewModel.ConvertMetricToModel(key, val)
		resmodelsTX = append(resmodelsTX, valNewModel)
		return nil
	})

	if err != nil {
		return err
	}
	h.l.L.Debug("Sending:", zap.Int("store len", len(resmodelsTX)))

	var b = 1
	if h.cfg.ContentBatch > 0 {
		b = int(h.cfg.ContentBatch)
	}

	if b > len(resmodelsTX) {
		b = len(resmodelsTX)
	}
	var x int
	needModels := make(map[job.JobID]bool)
	for ; x < len(resmodelsTX); x += b {
		if (x + b) > len(resmodelsTX) {
			b = len(resmodelsTX) - x
			if b == 0 {
				break
			}
		}
		id := h.pool.SendJob(ctx, resmodelsTX[x:x+b])
		h.l.L.Debug("SendedJob:", zap.Int("batch_len", b), zap.Int64("id", int64(id)))
		needModels[id] = true
	}

	var errRes error
	resultModels := make(map[job.JobID]job.Result, len(needModels))
	for range needModels {
		res := h.pool.GetResult(ctx)
		h.l.L.Debug("GetJob:", zap.Int64("id", int64(res.ID)))
		if res.Err != nil {
			errRes = res.Err
			h.l.L.Error("error result:", zap.Error(errRes))
		}
		resultModels[res.ID] = res
	}

	if errRes != nil {
		var rollBack []models.Metrics
		for _, v := range resmodelsTX {
			if v.MType == "counter" {
				rollBack = append(rollBack, v)
			}
		}
		if len(rollBack) > 0 {
			_, _ = h.store.AddMulti(ctx, rollBack)
			h.l.L.Debug("Rollback :", zap.Int("len", len(rollBack)))
		}
	}

	return errRes
}
