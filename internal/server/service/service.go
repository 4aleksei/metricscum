package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/store"
)

type serverMetricsStorage interface {
	Add(string, valuemetric.ValueMetric) valuemetric.ValueMetric
	Get(string) (valuemetric.ValueMetric, error)
	ReadAll(memstorage.FuncReadAllMetric) error
}

type HandlerStore struct {
	store serverMetricsStorage
	db    *store.DB
}

func NewHandlerStore(store serverMetricsStorage, db *store.DB) *HandlerStore {
	h := new(HandlerStore)
	h.store = store
	h.db = db
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

func (h *HandlerStore) SetValueModel(valModel models.Metrics) (*models.Metrics, error) {
	kind, errKind := valuemetric.GetKind(valModel.MType)
	if errKind != nil {
		return nil, fmt.Errorf("failed %w", errKind)
	}
	if valModel.ID == "" {
		return nil, fmt.Errorf("failed %w", ErrBadName)
	}
	val, err := valuemetric.ConvertToValueMetricInt(kind, valModel.Delta, valModel.Value)
	if err != nil {
		return nil, fmt.Errorf("failed %w", err)
	}
	newval := h.store.Add(valModel.ID, *val)
	valNewModel := new(models.Metrics)
	valNewModel.ConvertMetricToModel(valModel.ID, newval)
	return valNewModel, nil
}

func (h *HandlerStore) GetValueModel(valModel models.Metrics) (*models.Metrics, error) {
	kind, errKind := valuemetric.GetKind(valModel.MType)
	if errKind != nil {
		return nil, fmt.Errorf("failed %w", errKind)
	}
	if valModel.ID == "" {
		return nil, fmt.Errorf("failed %w", ErrBadName)
	}
	val, err := h.store.Get(valModel.ID)
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

func (h *HandlerStore) RecievePlainValue(typeVal, name, valstr string) error {
	kind, errKind := valuemetric.GetKind(typeVal)
	if errKind != nil {
		return fmt.Errorf("failed %w", errKind)
	}
	val, err := valuemetric.ConvertToValueMetric(kind, valstr)
	if err != nil {
		return fmt.Errorf("failed %w", err)
	}
	h.store.Add(name, *val)
	return nil
}

func (h *HandlerStore) GetValuePlain(name, typeVal string) (string, error) {
	val, err := h.store.Get(name)
	if err != nil {
		return "", fmt.Errorf("failed %w", err)
	}
	typeKind, errKind := valuemetric.GetKind(typeVal)
	if (errKind != nil) || (!val.KindOf(typeKind)) {
		return "", fmt.Errorf("failed %w", errKind)
	}
	_, valstr := valuemetric.ConvertValueMetricToPlain(val)
	return valstr, nil
}

func (h *HandlerStore) GetAllStore() (string, error) {
	var valstr string
	err := h.store.ReadAll(func(key string, val valuemetric.ValueMetric) error {
		_, value := valuemetric.ConvertValueMetricToPlain(val)
		valstr += fmt.Sprintf("%s : %s\n", key, value)
		return nil
	})
	if err != nil {
		return "", err
	}
	return valstr, nil
}

func (h *HandlerStore) GetPingDB(ctxPrnt context.Context) error {
	if h.db.DB == nil {
		return ErrNoDB
	}
	ctx, cancel := context.WithTimeout(ctxPrnt, 500*time.Millisecond)
	defer cancel()
	return h.db.DB.PingContext(ctx)
}
