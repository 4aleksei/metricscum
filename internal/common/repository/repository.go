package repository

import (
	"errors"
	"flag"
	"os"
	"strconv"
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

type Config struct {
	Interval int64
	Restore  bool
}

const WriteIntervalDefault int64 = 300
const RestoreDefault bool = true

func ReadConfigFlag(cfg *Config) {
	flag.Int64Var(&cfg.Interval, "i", WriteIntervalDefault, "Write data Interval")
	flag.BoolVar(&cfg.Restore, "r", RestoreDefault, "Restore data true/false")
}

func ReadConfigEnv(cfg *Config) {
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		val, err := strconv.Atoi(envStoreInterval)
		if err == nil {
			if val >= 0 {
				cfg.Interval = int64(val)
			}
		}
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		switch envRestore {
		case "true":
			cfg.Restore = true

		case "false":
			cfg.Restore = false
		}
	}
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

func (storage *MemStorageMuxLongTerm) doWriteData() {
	err := storage.filestorage.OpenWriter()
	if err != nil {
		logger.Log.Debug("error open source", zap.Error(err))
		return
	}
	defer func() { storage.filestorage.Close() }()
	valNewModel := new(models.Metrics)
	err = storage.store.ReadAll(func(key string, val valuemetric.ValueMetric) error {
		valNewModel.ConvertMetricToModel(key, val)
		if errson := storage.filestorage.WriteData(valNewModel); errson != nil {
			logger.Log.Debug("error writing data", zap.Error(errson))
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
}

func (storage *MemStorageMuxLongTerm) LoadData() error {
	err := storage.filestorage.OpenReader()
	if err != nil {
		logger.Log.Debug("error open source", zap.Error(err))
		return err
	}
	defer storage.filestorage.Close()

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
		_ = storage.store.Add(valNewModel.ID, *val)
	}
}

func (storage *MemStorageMuxLongTerm) saveData() {
	for {
		time.Sleep(time.Duration(storage.cfg.Interval) * time.Second)
		storage.DataWrite()
	}
}

func (storage *MemStorageMuxLongTerm) DataWrite() {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	storage.doWriteData()
}

func (storage *MemStorageMuxLongTerm) DataRun() {
	if storage.cfg.Restore {
		_ = storage.LoadData()
	}
	if storage.cfg.Interval > 0 {
		go storage.saveData()
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
