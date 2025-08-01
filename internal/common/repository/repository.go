// Package repository  - Storage base in file
package repository

import (
	"context"
	"errors"

	"fmt"

	"sync"
	"time"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"go.uber.org/zap"
)

type (
	longtermStorage interface {
		OpenWriter() error
		OpenReader() error
		WriteData(*models.Metrics) error
		ReadData(*models.Metrics) error
		CloseRead() error
		CloseWrite() error
	}

	Config struct {
		Interval int64
		Restore  bool
	}
)

type MemStorageMuxLongTerm struct {
	store       *memstorage.MemStorage
	mux         *sync.Mutex
	cfg         *Config
	filestorage longtermStorage
	l           *zap.Logger
}

func (storage *MemStorageMuxLongTerm) PingContext(ctx context.Context) error {
	return nil
}

func (storage *MemStorageMuxLongTerm) AddMulti(ctx context.Context, modval []models.Metrics) ([]models.Metrics, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	valNew, err := storage.store.AddMulti(ctx, modval)
	if err != nil {
		return nil, fmt.Errorf("failed add multi %w", err)
	}
	if storage.cfg.Interval == 0 {
		err := storage.doWriteData(ctx)
		if err != nil {
			storage.l.Error("error write data", zap.Error(err))
		}
	}
	return valNew, nil
}

func (storage *MemStorageMuxLongTerm) Add(ctx context.Context, name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	valNew, _ := storage.store.Add(ctx, name, val)
	if storage.cfg.Interval == 0 {
		err := storage.doWriteData(ctx)
		if err != nil {
			storage.l.Error("error write data", zap.Error(err))
		}
	}
	return valNew, nil
}

func (storage *MemStorageMuxLongTerm) Get(ctx context.Context, name string) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Get(ctx, name)
}

func (storage *MemStorageMuxLongTerm) ReadAll(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAll(ctx, prog)
}
func (storage *MemStorageMuxLongTerm) ReadAllClearCounters(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAllClearCounters(ctx, prog)
}

func (storage *MemStorageMuxLongTerm) doWriteData(ctx context.Context) error {
	err := storage.filestorage.OpenWriter()
	if err != nil {
		storage.l.Debug("error open source", zap.Error(err))
		return err
	}
	defer func() {
		if errF := storage.filestorage.CloseWrite(); errF != nil {
			storage.l.Debug("error writing data", zap.Error(errF))
		}
	}()

	valNewModel := new(models.Metrics)
	err = storage.store.ReadAll(ctx, func(key string, val valuemetric.ValueMetric) error {
		valNewModel.ConvertMetricToModel(key, val)
		if errson := storage.filestorage.WriteData(valNewModel); errson != nil {
			storage.l.Debug("error writing data", zap.Error(errson))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (storage *MemStorageMuxLongTerm) LoadData(ctx context.Context) error {
	err := storage.filestorage.OpenReader()
	if err != nil {
		storage.l.Debug("error open source", zap.Error(err))
		return err
	}
	defer func() {
		if err := storage.filestorage.CloseRead(); err != nil {
			storage.l.Debug("error writing data", zap.Error(err))
		}
	}()

	valNewModel := new(models.Metrics)
	for {
		if errson := storage.filestorage.ReadData(valNewModel); errson != nil {
			return errson
		}
		kind, errKind := valuemetric.GetKind(valNewModel.MType)
		if errKind != nil {
			return errKind
		}
		if valNewModel.ID == "" {
			return errors.New("no name")
		}
		val, err := valuemetric.ConvertToValueMetricInt(kind, valNewModel.Delta, valNewModel.Value)
		if err != nil {
			return err
		}
		_, _ = storage.store.Add(ctx, valNewModel.ID, *val)
	}
}

func (storage *MemStorageMuxLongTerm) saveData(ctx context.Context) {
	for {
		time.Sleep(time.Duration(storage.cfg.Interval) * time.Second)
		err := storage.DataWrite(ctx)
		if err != nil {
			storage.l.Error("error write data", zap.Error(err))
			continue
		}
	}
}

func (storage *MemStorageMuxLongTerm) DataWrite(ctx context.Context) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	return storage.doWriteData(ctx)
}

func (storage *MemStorageMuxLongTerm) DataRun(ctx context.Context) {
	if storage.cfg.Restore {
		_ = storage.LoadData(ctx)
	}
	if storage.cfg.Interval > 0 {
		go storage.saveData(ctx)
	}
}

func NewStoreMuxFiles(cfg *Config, l *zap.Logger, ltstore longtermStorage) *MemStorageMuxLongTerm {
	p := new(MemStorageMuxLongTerm)
	p.store = memstorage.NewStore()
	p.mux = new(sync.Mutex)
	p.cfg = cfg
	p.filestorage = ltstore
	p.l = l
	return p
}
