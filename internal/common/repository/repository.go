package repository

import (
	"errors"
	"strconv"
	"sync"
)

type GaugeMetric float64
type CounterMetric int64

var (
	ErrNotFoundName = errors.New("not found name")
)

type MetricsStorage interface {
	Update(string, GaugeMetric)
	Add(string, CounterMetric)
	GetGauge(string) (GaugeMetric, error)
	GetCounter(string) (CounterMetric, error)
	GetCounterAndClear(string) (CounterMetric, error)
	ReadAll(func(typename string, name string, value string) error) error
	ReadAllClearCounters(func(typename string, name string, value string) error) error
}

func AddGauge(storage MetricsStorage, name string, val GaugeMetric) {
	storage.Update(name, val)
}
func AddCounter(storage MetricsStorage, name string, val CounterMetric) {
	storage.Add(name, val)
}

func ReadAll(storage MetricsStorage, prog func(typename string, name string, value string) error) error {

	return storage.ReadAll(prog)
}

func ReadAllClearCounters(storage MetricsStorage, prog func(typename string, name string, value string) error) error {

	return storage.ReadAllClearCounters(prog)
}

func GetGauge(storage MetricsStorage, name string) (GaugeMetric, error) {
	return storage.GetGauge(name)
}
func GetCounter(storage MetricsStorage, name string) (CounterMetric, error) {
	return storage.GetCounter(name)
}

func GetCounterAndClear(storage MetricsStorage, name string) (CounterMetric, error) {
	return storage.GetCounterAndClear(name)
}

type MemStorage struct {
	gauge   map[string]GaugeMetric
	counter map[string]CounterMetric
}

type MemStorageMux struct {
	store *MemStorage
	mux   *sync.Mutex
}

func (storage *MemStorage) Update(name string, val GaugeMetric) {
	storage.gauge[name] = val
}

func (storage *MemStorage) Add(name string, val CounterMetric) {
	storage.counter[name] += val
}

func (storage *MemStorageMux) Update(name string, val GaugeMetric) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	storage.store.gauge[name] = val
}

func (storage *MemStorageMux) Add(name string, val CounterMetric) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	storage.store.counter[name] += val
}

func (storage *MemStorage) GetCounter(name string) (CounterMetric, error) {

	if val, ok := storage.counter[name]; ok {
		return val, nil
	} else {
		return val, ErrNotFoundName
	}

}

func (storage *MemStorage) GetCounterAndClear(name string) (CounterMetric, error) {

	if val, ok := storage.counter[name]; ok {
		storage.counter[name] = 0
		return val, nil
	} else {
		return val, ErrNotFoundName
	}
}

func (storage *MemStorage) GetGauge(name string) (GaugeMetric, error) {

	if val, ok := storage.gauge[name]; ok {
		return val, nil
	} else {
		return val, ErrNotFoundName
	}

}

func (storage *MemStorageMux) GetCounter(name string) (CounterMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.GetCounter(name)
}

func (storage *MemStorageMux) GetCounterAndClear(name string) (CounterMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	return storage.store.GetCounterAndClear(name)
}

func (storage *MemStorageMux) GetGauge(name string) (GaugeMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.GetGauge(name)
}

func (storage *MemStorageMux) ReadAll(prog func(typename string, name string, value string) error) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	return storage.store.ReadAll(prog)

}

func (storage *MemStorageMux) ReadAllClearCounters(prog func(typename string, name string, value string) error) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	return storage.store.ReadAllClearCounters(prog)

}

func (storage *MemStorage) ReadAllClearCounters(prog func(typename string, name string, value string) error) error {

	for name, gauge := range storage.gauge {
		valstr := strconv.FormatFloat(float64(gauge), 'f', -1, 64)

		err := prog("gauge", name, valstr)
		if err != nil {
			return err
		}
	}

	for name, counter := range storage.counter {

		valstr := strconv.FormatInt(int64(counter), 10)
		err := prog("counter", name, valstr)
		if err != nil {
			return err
		} else {
			storage.counter[name] = 0
		}

	}

	return nil
}

func (storage *MemStorage) ReadAll(prog func(typename string, name string, value string) error) error {

	for name, gauge := range storage.gauge {
		valstr := strconv.FormatFloat(float64(gauge), 'f', -1, 64)

		err := prog("gauge", name, valstr)
		if err != nil {
			return err
		}
	}

	for name, counter := range storage.counter {

		valstr := strconv.FormatInt(int64(counter), 10)
		err := prog("counter", name, valstr)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewStore() *MemStorage {
	return &MemStorage{
		counter: make(map[string]CounterMetric),
		gauge:   make(map[string]GaugeMetric),
	}
}

func NewStoreMux() *MemStorageMux {
	return &MemStorageMux{
		store: NewStore(),
		mux:   &sync.Mutex{},
	}
}
