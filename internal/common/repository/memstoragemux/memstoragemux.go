// Package memstoragemux
package memstoragemux

import (
	"context"
	"sync"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

type MemStorageMux struct {
	store *memstorage.MemStorage
	mux   *sync.Mutex
}

func (storage *MemStorageMux) PingContext(ctx context.Context) error {
	return nil
}

func (storage *MemStorageMux) AddMulti(ctx context.Context, modval []models.Metrics) ([]models.Metrics, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.AddMulti(ctx, modval)
}

func (storage *MemStorageMux) Add(ctx context.Context, name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Add(ctx, name, val)
}

func (storage *MemStorageMux) Get(ctx context.Context, name string) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Get(ctx, name)
}

func (storage *MemStorageMux) ReadAll(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAll(ctx, prog)
}
func (storage *MemStorageMux) ReadAllClearCounters(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAllClearCounters(ctx, prog)
}

func NewStoreMux() *MemStorageMux {
	p := new(MemStorageMux)
	p.store = memstorage.NewStore()
	p.mux = new(sync.Mutex)
	return p
}
