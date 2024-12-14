package repository

import (
	"errors"
	"strconv"
	"sync"
)

type GaugeMetric float64
type CounterMetric int64

type valueKind int

const (
	kindBadEmpty valueKind = iota
	kindInt64
	kindFloat64
)

type valueMetric struct {
	kind       valueKind
	valueFloat GaugeMetric
	valueInt   CounterMetric
}

type FuncReadAllMetric func(typename string, name string, value string) error

var (
	ErrNotFoundName = errors.New("not found name")
)

type MemStorage struct {
	values map[string]valueMetric
}

type MemStorageMux struct {
	store *MemStorage
	mux   *sync.Mutex
}

func (storage *MemStorage) Update(name string, val GaugeMetric) {
	if entry, ok := storage.values[name]; ok {
		entry.kind = kindFloat64
		entry.valueFloat = val
		storage.values[name] = entry
	} else {
		storage.values[name] = valueMetric{kind: kindFloat64, valueFloat: val}
	}
}

func (storage *MemStorage) Add(name string, val CounterMetric) {
	if entry, ok := storage.values[name]; ok {
		entry.kind = kindInt64
		entry.valueInt += val
		storage.values[name] = entry
	} else {
		storage.values[name] = valueMetric{kind: kindInt64, valueInt: val}
	}
}
func (storage *MemStorage) GetCounter(name string) (CounterMetric, error) {
	if entry, ok := storage.values[name]; ok {
		if entry.kind == kindInt64 {
			return entry.valueInt, nil
		}
	}
	return 0, ErrNotFoundName
}

func (storage *MemStorage) GetGauge(name string) (GaugeMetric, error) {
	if entry, ok := storage.values[name]; ok {
		if entry.kind == kindFloat64 {
			return entry.valueFloat, nil
		}
	}
	return 0, ErrNotFoundName
}

func (storage *MemStorageMux) Update(name string, val GaugeMetric) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	storage.store.Update(name, val)
}

func (storage *MemStorageMux) Add(name string, val CounterMetric) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	storage.store.Add(name, val)
}

func (storage *MemStorageMux) GetCounter(name string) (CounterMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.GetCounter(name)
}

func (storage *MemStorageMux) GetGauge(name string) (GaugeMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.GetGauge(name)
}

func (storage *MemStorageMux) ReadAll(prog FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAll(prog)
}
func (storage *MemStorageMux) ReadAllClearCounters(prog FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAllClearCounters(prog)
}
func (storage *MemStorage) ReadAllClearCounters(prog FuncReadAllMetric) error {
	for name, val := range storage.values {
		var valstr string
		var ty string
		switch val.kind {
		case kindFloat64:
			valstr = strconv.FormatFloat(float64(val.valueFloat), 'f', -1, 64)
			ty = "gauge"
		case kindInt64:
			valstr = strconv.FormatInt(int64(val.valueInt), 10)
			ty = "counter"
		default:
			continue
		}
		err := prog(ty, name, valstr)
		if err != nil {
			return err
		}
		if val.kind == kindInt64 {
			val.valueInt = 0
			storage.values[name] = val
		}
	}
	return nil
}

func (storage *MemStorage) ReadAll(prog FuncReadAllMetric) error {
	for name, val := range storage.values {
		var valstr string
		var ty string
		switch val.kind {
		case kindFloat64:
			valstr = strconv.FormatFloat(float64(val.valueFloat), 'f', -1, 64)
			ty = "gauge"
		case kindInt64:
			valstr = strconv.FormatInt(int64(val.valueInt), 10)
			ty = "counter"
		default:
			continue
		}
		err := prog(ty, name, valstr)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewStore() *MemStorage {
	p := new(MemStorage)
	p.values = make(map[string]valueMetric)
	return p
}

func NewStoreMux() *MemStorageMux {
	p := new(MemStorageMux)
	p.store = NewStore()
	p.mux = new(sync.Mutex)
	return p
}
