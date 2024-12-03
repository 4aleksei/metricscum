package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/4aleksei/metricscum/internal/common/repository"
)

type HandlerStore struct {
	Store repository.MetricsStorage
}

func NewHandlerStore(store repository.MetricsStorage) *HandlerStore {
	return &HandlerStore{
		Store: store,
	}
}

var (
	ErrBadValue = errors.New("invalid value")
)

func RecieveGauge(store repository.MetricsStorage, name string, valstr string) error {

	value, err := strconv.ParseFloat(valstr, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}

	repository.AddGauge(store, name, repository.GaugeMetric(value))
	return nil
}

func RecieveCounter(store repository.MetricsStorage, name string, valstr string) error {
	value, err := strconv.ParseInt(valstr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed %w : %w", ErrBadValue, err)
	}
	repository.AddCounter(store, name, repository.CounterMetric(value))
	return nil
}
