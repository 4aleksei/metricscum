package memstoragemux

import (
	"context"
	"sync"

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

func (storage *MemStorageMux) Add(name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Add(name, val)
}

func (storage *MemStorageMux) Get(name string) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Get(name)
}

func (storage *MemStorageMux) ReadAll(prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAll(prog)
}
func (storage *MemStorageMux) ReadAllClearCounters(prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAllClearCounters(prog)
}

func NewStoreMux() *MemStorageMux {
	p := new(MemStorageMux)
	p.store = memstorage.NewStore()
	p.mux = new(sync.Mutex)
	return p
}
