package service

import (
	"errors"
	"fmt"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository"
)

type ServerMetricsStorage interface {
	Add(string, repository.ValueMetric) repository.ValueMetric
	Get(string) (repository.ValueMetric, error)
	ReadAll(repository.FuncReadAllMetric) error
}

type HandlerStore struct {
	store ServerMetricsStorage
}

func NewHandlerStore(store ServerMetricsStorage) *HandlerStore {
	h := new(HandlerStore)
	h.store = store
	return h
}

var (
	ErrBadValue = errors.New("invalid value")
	ErrBadName  = errors.New("no name")
)

func (h *HandlerStore) CheckType(s string) error {
	_, errKind := repository.GetKind(s)
	if errKind != nil {
		return errKind
	}
	return nil
}

func (h *HandlerStore) SetValueModel(valModel models.Metrics) (*models.Metrics, error) {
	kind, errKind := repository.GetKind(valModel.MType)
	if errKind != nil {
		return nil, fmt.Errorf("failed %w", errKind)
	}
	if valModel.ID == "" {
		return nil, fmt.Errorf("failed %w", ErrBadName)
	}
	val, err := repository.ConvertToValueMetricInt(kind, valModel.Delta, valModel.Value)
	if err != nil {
		return nil, fmt.Errorf("failed %w", err)
	}
	newval := h.store.Add(valModel.ID, *val)

	valNewModel := new(models.Metrics)
	valNewModel.ConvertMetricToModel(valModel.ID, newval)
	return valNewModel, nil
}

func (h *HandlerStore) GetValueModel(valModel models.Metrics) (*models.Metrics, error) {
	kind, errKind := repository.GetKind(valModel.MType)
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

func (h *HandlerStore) RecievePlainValue(typeVal string, name string, valstr string) error {
	kind, errKind := repository.GetKind(typeVal)
	if errKind != nil {
		return fmt.Errorf("failed %w", errKind)
	}
	val, err := repository.ConvertToValueMetric(kind, valstr)

	if err != nil {
		return fmt.Errorf("failed %w", err)
	}
	h.store.Add(name, *val)
	return nil
}

func (h *HandlerStore) GetValuePlain(name string, typeVal string) (string, error) {
	val, err := h.store.Get(name)
	if err != nil {
		return "", fmt.Errorf("failed %w", err)
	}
	typeKind, errKind := repository.GetKind(typeVal)
	if (errKind != nil) || (!val.KindOf(typeKind)) {
		return "", fmt.Errorf("failed %w", errKind)
	}
	_, valstr := repository.ConvertValueMetricToPlain(val)
	return valstr, nil
}

func (h *HandlerStore) GetAllStore() (string, error) {

	var valstr string

	err := h.store.ReadAll(func(key string, val repository.ValueMetric) error {
		_, value := repository.ConvertValueMetricToPlain(val)
		valstr += (key + " : " + value + "\n")

		return nil
	})

	if err != nil {
		return "", err
	}

	return valstr, nil
}
