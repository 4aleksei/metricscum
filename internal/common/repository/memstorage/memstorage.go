package memstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

type MemStorage struct {
	values map[string]valuemetric.ValueMetric
}

type FuncReadAllMetric func(name string, val valuemetric.ValueMetric) error

var (
	ErrNotFoundName = errors.New("not found name")
)

func (storage *MemStorage) PingContext(ctx context.Context) error {
	return nil
}

var (
	ErrBadValue = errors.New("invalid value")
	ErrBadName  = errors.New("no name")
	ErrNoDB     = errors.New("no db")
)

func (storage *MemStorage) AddMulti(modval []models.Metrics) (*[]models.Metrics, error) {
	resmodels := new([]models.Metrics)
	for _, valModel := range modval {
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
		resval, errA := storage.Add(valModel.ID, *val)
		if errA != nil {
			return nil, errA
		}
		var valNewModel models.Metrics
		valNewModel.ConvertMetricToModel(valModel.ID, resval)
		*resmodels = append(*resmodels, valNewModel)
	}
	return resmodels, nil
}

func (storage *MemStorage) Add(name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
	if entry, ok := storage.values[name]; ok {
		entry.DoUpdate(val)
		storage.values[name] = entry
		return entry, nil
	}
	storage.values[name] = val
	return val, nil
}

func (storage *MemStorage) Get(name string) (valuemetric.ValueMetric, error) {
	if entry, ok := storage.values[name]; ok {
		return entry, nil
	}
	return valuemetric.ValueMetric{}, ErrNotFoundName
}

func (storage *MemStorage) ReadAllClearCounters(prog FuncReadAllMetric) error {
	for name, entry := range storage.values {
		err := prog(name, entry)
		if err != nil {
			return err
		}
		storage.values[name] = entry.DoRead()
	}
	return nil
}

func (storage *MemStorage) ReadAll(prog FuncReadAllMetric) error {
	for name, entry := range storage.values {
		err := prog(name, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewStore() *MemStorage {
	p := new(MemStorage)
	p.values = make(map[string]valuemetric.ValueMetric)
	return p
}
