package repository

import (
	"strconv"
	"sync"
)

type GaugeMetric float64
type CounterMetric int64

type MetricsStorage interface {
	Update(string, GaugeMetric)
	Add(string, CounterMetric)
	GetGauge(string) GaugeMetric
	GetCounter(string) CounterMetric
	GetCounterAndClear(string) CounterMetric
	ReadAll(func(typename string, name string, value string) error) error
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

func GetGauge(storage MetricsStorage, name string) GaugeMetric {
	return storage.GetGauge(name)
}
func GetCounter(storage MetricsStorage, name string) CounterMetric {
	return storage.GetCounter(name)
}

func GetCounterAndClear(storage MetricsStorage, name string) CounterMetric {
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

func (storage *MemStorage) GetCounter(name string) CounterMetric {

	return storage.counter[name]
}

func (storage *MemStorage) GetCounterAndClear(name string) CounterMetric {
	val := storage.counter[name]
	storage.counter[name] = 0
	return val
}

func (storage *MemStorage) GetGauge(name string) GaugeMetric {
	return storage.gauge[name]
}

func (storage *MemStorageMux) GetCounter(name string) CounterMetric {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.counter[name]
}

func (storage *MemStorageMux) GetCounterAndClear(name string) CounterMetric {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	val := storage.store.counter[name]
	storage.store.counter[name] = 0
	return val
}

func (storage *MemStorageMux) GetGauge(name string) GaugeMetric {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.gauge[name]
}

func (storage *MemStorageMux) ReadAll(prog func(typename string, name string, value string) error) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	return storage.store.ReadAll(prog)

}

func (storage *MemStorage) ReadAll(prog func(typename string, name string, value string) error) error {

	for name, gauge := range storage.gauge {
		valstr := strconv.FormatFloat(float64(gauge), 'E', -1, 64)

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
