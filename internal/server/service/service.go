package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/4aleksei/metricscum/internal/common/repository"
)

type ServerMetricsStorage interface {
	Update(string, repository.GaugeMetric)
	Add(string, repository.CounterMetric)
	GetGauge(string) (repository.GaugeMetric, error)
	GetCounter(string) (repository.CounterMetric, error)
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
)

func (h *HandlerStore) RecieveGauge(name string, valstr string) error {

	value, err := strconv.ParseFloat(valstr, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}

	h.store.Update(name, repository.GaugeMetric(value))
	return nil
}

func (h *HandlerStore) RecieveCounter(name string, valstr string) error {
	value, err := strconv.ParseInt(valstr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	h.store.Add(name, repository.CounterMetric(value))
	return nil
}

func (h *HandlerStore) GetGauge(name string) (string, error) {

	val, err := h.store.GetGauge(name)

	if err != nil {
		return "", err
	}
	valstr := strconv.FormatFloat(float64(val), 'f', -1, 64)
	return valstr, nil
}

func (h *HandlerStore) GetCounter(name string) (string, error) {

	val, err := h.store.GetCounter(name)

	if err != nil {
		return "", err
	}
	valstr := strconv.FormatInt(int64(val), 10)
	return valstr, nil
}

func (h *HandlerStore) GetAllStore() (string, error) {

	var valstr string

	err := h.store.ReadAll(func(typename string, name string, value string) error {

		valstr += (name + " : " + value + "\n")

		return nil
	})

	if err != nil {
		return "", err
	}

	return valstr, nil
}
