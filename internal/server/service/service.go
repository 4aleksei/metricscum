// Package service - Metrics server service
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

type serverMetricsStorage interface {
	Add(context.Context, string, valuemetric.ValueMetric) (valuemetric.ValueMetric, error)
	Get(context.Context, string) (valuemetric.ValueMetric, error)
	ReadAll(context.Context, memstorage.FuncReadAllMetric) error
	PingContext(context.Context) error
	AddMulti(context.Context, []models.Metrics) ([]models.Metrics, error)
}

type HandlerStore struct {
	store serverMetricsStorage
}

func NewHandlerStore(store serverMetricsStorage) *HandlerStore {
	h := new(HandlerStore)
	h.store = store
	return h
}

var (
	ErrBadValue = errors.New("invalid value")
	ErrBadName  = errors.New("no name")
	ErrNoDB     = errors.New("no db")
)

func (h *HandlerStore) CheckType(s string) error {
	_, errKind := valuemetric.GetKind(s)
	if errKind != nil {
		return errKind
	}
	return nil
}

func (h *HandlerStore) SetValueSModel(ctx context.Context, valModel []models.Metrics) ([]models.Metrics, error) {
	valNewModel, errA := h.store.AddMulti(ctx, valModel)
	if errA != nil {
		return nil, fmt.Errorf("add failed %w", errA)
	}
	return valNewModel, nil
}

func (h *HandlerStore) SetValueModel(ctx context.Context, valModel models.Metrics) (*models.Metrics, error) {
	kind, errKind := valuemetric.GetKind(valModel.MType)
	if errKind != nil {
		return nil, fmt.Errorf("failed kind %w", errKind)
	}
	if valModel.ID == "" {
		return nil, fmt.Errorf("failed %w", ErrBadName)
	}
	val, err := valuemetric.ConvertToValueMetricInt(kind, valModel.Delta, valModel.Value)
	if err != nil {
		return nil, fmt.Errorf("failed %w", err)
	}
	newval, errA := h.store.Add(ctx, valModel.ID, *val)
	if errA != nil {
		return nil, fmt.Errorf("add failed %w", errA)
	}

	valNewModel := new(models.Metrics)
	valNewModel.ConvertMetricToModel(valModel.ID, newval)
	return valNewModel, nil
}

func (h *HandlerStore) GetValueModel(ctx context.Context, valModel models.Metrics) (*models.Metrics, error) {
	kind, errKind := valuemetric.GetKind(valModel.MType)
	if errKind != nil {
		return nil, fmt.Errorf("kind failed %w", errKind)
	}
	if valModel.ID == "" {
		return nil, fmt.Errorf("failed %w", ErrBadName)
	}
	val, err := h.store.Get(ctx, valModel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed %w", err)
	}
	if !val.KindOf(kind) {
		return nil, fmt.Errorf("failed %w", err)
	}
	valNewModel := new(models.Metrics)
	valNewModel.ConvertMetricToModel(valModel.ID, val)
	return valNewModel, nil
}

func (h *HandlerStore) RecievePlainValue(ctx context.Context, typeVal, name, valstr string) error {
	kind, errKind := valuemetric.GetKind(typeVal)
	if errKind != nil {
		return fmt.Errorf("kind failed %w", errKind)
	}
	val, err := valuemetric.ConvertToValueMetric(kind, valstr)
	if err != nil {
		return fmt.Errorf("failed %w", err)
	}
	_, err = h.store.Add(ctx, name, *val)
	if err != nil {
		return fmt.Errorf("failed %w", err)
	}
	return nil
}

func (h *HandlerStore) GetValuePlain(ctx context.Context, name, typeVal string) (string, error) {
	val, err := h.store.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed %w", err)
	}
	typeKind, errKind := valuemetric.GetKind(typeVal)

	if (errKind != nil) || (!val.KindOf(typeKind)) {
		return "", fmt.Errorf("kind failed %w", errKind)
	}
	_, valstr := valuemetric.ConvertValueMetricToPlain(val)
	return valstr, nil
}

func (h *HandlerStore) GetAllStore(ctx context.Context) (string, error) {
	var valstr string
	err := h.store.ReadAll(ctx, func(key string, val valuemetric.ValueMetric) error {
		_, value := valuemetric.ConvertValueMetricToPlain(val)
		valstr += fmt.Sprintf("%s : %s\n", key, value)
		return nil
	})
	if err != nil {
		return "", err
	}
	return valstr, nil
}

func (h *HandlerStore) GetPingDB(ctx context.Context) error {
	return h.store.PingContext(ctx)
}
