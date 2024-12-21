package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"go.uber.org/zap"
)

type LongtermStorage interface {
	OpenWriter() error
	OpenReader() error
	WriteData(*models.Metrics) error
	ReadData(*models.Metrics) error
	Close() error
}

// filePath string
type Config struct {
	Interval uint
	Restore  bool
}

type MemStorageMuxLongTerm struct {
	store       *memstorage.MemStorage
	mux         *sync.Mutex
	cfg         *Config
	filestorage LongtermStorage
}

func (storage *MemStorageMuxLongTerm) Add(name string, val valuemetric.ValueMetric) valuemetric.ValueMetric {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	valNew := storage.store.Add(name, val)
	if storage.cfg.Interval == 0 {
		storage.doWriteData()
	}
	return valNew
}

func (storage *MemStorageMuxLongTerm) Get(name string) (valuemetric.ValueMetric, error) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.Get(name)
}

func (storage *MemStorageMuxLongTerm) ReadAll(prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAll(prog)
}
func (storage *MemStorageMuxLongTerm) ReadAllClearCounters(prog memstorage.FuncReadAllMetric) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	return storage.store.ReadAllClearCounters(prog)
}

func (store *MemStorageMuxLongTerm) doWriteData() {
	err := store.filestorage.OpenWriter()
	if err != nil {
		logger.Log.Debug("error open source", zap.Error(err))
		return
	}
	defer func() { store.filestorage.Close() }()
	valNewModel := new(models.Metrics)
	err = store.store.ReadAll(func(key string, val valuemetric.ValueMetric) error {
		valNewModel.ConvertMetricToModel(key, val)

		if errson := store.filestorage.WriteData(valNewModel); errson != nil {
			logger.Log.Debug("error writing data", zap.Error(errson))
			return err
		}

		return nil
	})
	if err != nil {
		return
	}
}

func (store *MemStorageMuxLongTerm) LoadData() error {

	err := store.filestorage.OpenReader()
	if err != nil {
		logger.Log.Debug("error open source", zap.Error(err))
		return err
	}
	defer store.filestorage.Close()

	valNewModel := new(models.Metrics)
	for {
		if errson := store.filestorage.ReadData(valNewModel); errson != nil {
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
		_ = store.store.Add(valNewModel.ID, *val)
	}
}

func (store *MemStorageMuxLongTerm) saveData() {
	for {
		time.Sleep(time.Duration(store.cfg.Interval) * time.Second)
		store.DataWrite()
	}
}

func (store *MemStorageMuxLongTerm) DataWrite() {
	store.mux.Lock()
	store.doWriteData()
	store.mux.Unlock()
}

func (store *MemStorageMuxLongTerm) DataRun() {
	if store.cfg.Restore {
		store.LoadData()
	}
	if store.cfg.Interval > 0 {
		go store.saveData()
	}
}

func NewStoreMuxFiles(cfg *Config, ltstore LongtermStorage) *MemStorageMuxLongTerm {
	p := new(MemStorageMuxLongTerm)
	p.store = memstorage.NewStore()
	p.mux = new(sync.Mutex)
	p.cfg = cfg
	p.filestorage = ltstore
	return p
}
