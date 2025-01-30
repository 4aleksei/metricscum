package service

import (
	"context"
	"errors"
	"time"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/utils"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool"
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

const (
	waitGroupSleep int = 10
)

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

func (h *HandlerStore) rollBackMetrics(ctx context.Context, resmodelsTX []models.Metrics) {
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

func (h *HandlerStore) sendMetricsRun(ctx context.Context, resmodelsTX []models.Metrics, wgRes *utils.WaitGroupTimeout) error {
	var b = 1
	if h.cfg.ContentBatch > 0 {
		b = int(h.cfg.ContentBatch)
	}
	if b > len(resmodelsTX) {
		b = len(resmodelsTX)
	}
	for x := 0; x < len(resmodelsTX); x += b {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if (x + b) > len(resmodelsTX) {
				b = len(resmodelsTX) - x
				if b == 0 {
					break
				}
			}
			wgRes.Add(1)
			id := h.pool.SendJob(ctx, resmodelsTX[x:x+b])
			h.l.L.Debug("SendedJob:", zap.Int("batch_len", b), zap.Int64("id", int64(id)))
		}
	}
	return nil
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

	var errRes error
	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	wgRes := &utils.WaitGroupTimeout{}

	go func(ctx context.Context, wg *utils.WaitGroupTimeout) {
		for {
			select {
			case <-ctx.Done():
				h.l.L.Debug("Exit routin:")
				return
			default:
				res, err := h.pool.GetResult(ctx)
				if err != nil {
					h.l.L.Debug("Done:", zap.Error(err))
					return
				}
				h.l.L.Debug("GetJob:", zap.Int64("id", int64(res.ID)))
				wg.Done()

				if res.Err != nil {
					errRes = res.Err
					h.l.L.Error("error result:", zap.Error(errRes))
				}
			}
		}
	}(ctxCancel, wgRes)

	_ = h.sendMetricsRun(ctx, resmodelsTX, wgRes)

	for {
		err := wgRes.WaitWithTimeout(ctx, time.Duration(waitGroupSleep)*time.Second)
		if err != nil {
			if errors.Is(err, utils.ErrWgWaitTimeOut) {
				continue
			}
			return err
		}
		break
	}
	if errRes != nil {
		h.rollBackMetrics(ctx, resmodelsTX)
	} else {
		h.l.L.Debug("Sending success", zap.Int("len", len(resmodelsTX)))
	}
	return errRes
}
