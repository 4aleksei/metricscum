package repository

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

type valueKind int

const (
	kindBadEmpty valueKind = iota
	kindInt64
	kindFloat64
)

type ValueMetric struct {
	kind       valueKind
	valueFloat float64
	valueInt   int64
}

type FuncReadAllMetric func(name string, val ValueMetric) error

var (
	ErrNotFoundName = errors.New("not found name")
)

func (v *ValueMetric) GetTypeStr() string {
	return GetKindStr(v.kind)
}

func (v *ValueMetric) ValueInt() *int64 {
	if v.kind == kindInt64 {
		return &v.valueInt
	}
	return nil
}

func (v *ValueMetric) ValueFloat() *float64 {
	if v.kind == kindFloat64 {
		return &v.valueFloat
	}
	return nil
}

func (v *ValueMetric) doUpdate(val ValueMetric) {
	switch v.kind {
	case kindFloat64:
		v.valueFloat = val.valueFloat
	case kindInt64:
		v.valueInt += val.valueInt
	default:
	}
}

func (v *ValueMetric) doRead() ValueMetric {
	switch v.kind {
	case kindInt64:
		v.valueInt = 0
	default:
	}
	return *v
}

func (v *ValueMetric) KindOf(k valueKind) bool {
	return v.kind == k
}

var (
	ErrBadTypeValue = errors.New("invalid typeValue")
	ErrBadValue     = errors.New("error value conversion")
	ErrBadKindType  = errors.New("error kind type")
)

func GetKind(typeValue string) (valueKind, error) {
	switch typeValue {
	case "gauge":
		return kindFloat64, nil
	case "counter":
		return kindInt64, nil
	default:
		return kindBadEmpty, ErrBadTypeValue
	}
}

func GetKindStr(typeValue valueKind) string {
	switch typeValue {
	case kindFloat64:
		return "gauge"
	case kindInt64:
		return "counter"
	default:
		return ""
	}
}

func ConvertToFloatValueMetric(valF float64) *ValueMetric {
	val := new(ValueMetric)
	val.kind = kindFloat64
	val.valueFloat = valF
	return val
}

func ConvertToIntValueMetric(valI int64) *ValueMetric {
	val := new(ValueMetric)
	val.kind = kindInt64
	val.valueInt = valI
	return val
}

func ConvertToValueMetricInt(kind valueKind, delta *int64, value *float64) (*ValueMetric, error) {
	val := new(ValueMetric)
	val.kind = kind
	var err error
	switch kind {
	case kindFloat64:
		if value == nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}
		val.valueFloat = *value

	case kindInt64:
		if delta == nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}
		val.valueInt = *delta

	default:
		return nil, fmt.Errorf("failed %w : %w", ErrBadValue, ErrBadKindType)
	}
	return val, nil
}

func ConvertToValueMetric(kind valueKind, valstr string) (*ValueMetric, error) {
	val := new(ValueMetric)
	val.kind = kind
	var err error
	switch kind {
	case kindFloat64:
		val.valueFloat, err = strconv.ParseFloat(valstr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}

	case kindInt64:
		val.valueInt, err = strconv.ParseInt(valstr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}

	default:
		return nil, fmt.Errorf("failed %w : %w", ErrBadValue, ErrBadKindType)
	}
	return val, nil
}

func ConvertValueMetricToPlain(val ValueMetric) (a, b string) {
	switch val.kind {
	case kindFloat64:
		a = GetKindStr(val.kind)
		b = strconv.FormatFloat(val.valueFloat, 'f', -1, 64)
	case kindInt64:
		a = GetKindStr(val.kind)
		b = strconv.FormatInt(val.valueInt, 10)
	}
	return a, b
}

type MemStorage struct {
	values map[string]ValueMetric
}

type MemStorageMux struct {
	store *MemStorage
	mux   *sync.Mutex
}

func (storage *MemStorage) Add(name string, val ValueMetric) ValueMetric {
	if entry, ok := storage.values[name]; ok {
		entry.doUpdate(val)
		storage.values[name] = entry
		return entry
	}
	storage.values[name] = val
	return val
}

func (storage *MemStorage) Get(name string) (ValueMetric, error) {
	if entry, ok := storage.values[name]; ok {
		return entry, nil
	}
	return ValueMetric{}, ErrNotFoundName
}

func (storage *MemStorageMux) Add(name string, val ValueMetric) ValueMetric {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Add(name, val)
}

func (storage *MemStorageMux) Get(name string) (ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Get(name)
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
	for name, entry := range storage.values {
		err := prog(name, entry)
		if err != nil {
			return err
		}
		storage.values[name] = entry.doRead()
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
	p.values = make(map[string]ValueMetric)
	return p
}

func NewStoreMux() *MemStorageMux {
	p := new(MemStorageMux)
	p.store = NewStore()
	p.mux = new(sync.Mutex)
	return p
}
