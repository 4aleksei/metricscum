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

func GetGauge(store repository.MetricsStorage, name string) (string, error) {

	val, err := repository.GetGauge(store, name)

	if err != nil {
		return "", err
	}
	valstr := strconv.FormatFloat(float64(val), 'f', -1, 64)
	return valstr, nil
}

func GetCounter(store repository.MetricsStorage, name string) (string, error) {

	val, err := repository.GetCounter(store, name)

	if err != nil {
		return "", err
	}
	valstr := strconv.FormatInt(int64(val), 10)
	return valstr, nil
}

func GetAllStore(store repository.MetricsStorage) (string, error) {

	var valstr string

	err := store.ReadAll(func(typename string, name string, value string) error {

		valstr += (name + " : " + value + "\n")

		return nil
	})

	if err != nil {
		return "", err
	}

	return valstr, nil
}
