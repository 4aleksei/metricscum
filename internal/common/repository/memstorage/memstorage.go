package memstorage

import (
	"errors"

	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

type MemStorage struct {
	values map[string]valuemetric.ValueMetric
}

type FuncReadAllMetric func(name string, val valuemetric.ValueMetric) error

var (
	ErrNotFoundName = errors.New("not found name")
)

func (storage *MemStorage) Add(name string, val valuemetric.ValueMetric) valuemetric.ValueMetric {
	if entry, ok := storage.values[name]; ok {
		entry.DoUpdate(val)
		storage.values[name] = entry
		return entry
	}
	storage.values[name] = val
	return val
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
