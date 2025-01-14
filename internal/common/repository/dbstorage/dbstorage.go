package dbstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/store"
	"go.uber.org/zap"
)

type DBStorage struct {
	db *store.DB
	l  *zap.Logger
}

var (
	ErrBadValue = errors.New("invalid value")
	ErrBadName  = errors.New("no name")
	ErrNoDB     = errors.New("no db")
)

func NewStoreDB(db *store.DB, l *zap.Logger) *DBStorage {
	return &DBStorage{db: db,
		l: l}
}

func (storage *DBStorage) PingContext(ctx context.Context) error {
	return storage.db.DB.PingContext(ctx)
}

func (storage *DBStorage) AddMulti(modval []models.Metrics) (*[]models.Metrics, error) {
	tx, err := storage.db.BeginTx()

	if err != nil {
		storage.l.Error("failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed begin tx %w", err)
	}

	resmodels := new([]models.Metrics)
	for _, valModel := range modval {
		kind, errKind := valuemetric.GetKind(valModel.MType)
		if errKind != nil {
			return nil, fmt.Errorf("failed kind %w", errKind)
		}
		if valModel.ID == "" {
			return nil, fmt.Errorf("failed %w", ErrBadName)
		}

		var valret *valuemetric.ValueMetric
		err = tx.Upsert(valModel.ID, int(kind), valModel.Delta, valModel.Value, func(n string, k int, d int64, v float64) error {
			kind, errK := valuemetric.GetKindInt(k)
			if errK != nil {
				return errK
			}
			valret, errK = valuemetric.ConvertToValueMetricInt(kind, &d, &v)
			return errK
		})
		if err != nil {
			storage.l.Error("failed to upsert transaction", zap.Error(err))
			return nil, err
		}
		var valNewModel models.Metrics
		valNewModel.ConvertMetricToModel(valModel.ID, *valret)
		*resmodels = append(*resmodels, valNewModel)
	}
	err = tx.EndTx()
	if err != nil {
		storage.l.Error("failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("failed end tx %w", err)
	}
	return resmodels, nil
}

func (storage *DBStorage) Add(name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
	tx, err := storage.db.BeginTx()
	if err != nil {
		storage.l.Error("failed to begin transaction", zap.Error(err))
		return valuemetric.ValueMetric{}, err
	}
	var valret *valuemetric.ValueMetric
	err = tx.Upsert(name, val.GetKind(), val.ValueInt(), val.ValueFloat(), func(n string, k int, d int64, v float64) error {
		kind, errK := valuemetric.GetKindInt(k)
		if errK != nil {
			return errK
		}
		valret, errK = valuemetric.ConvertToValueMetricInt(kind, &d, &v)
		return errK
	})
	if err != nil {
		storage.l.Error("failed to upsert transaction", zap.Error(err))
		return valuemetric.ValueMetric{}, err
	}
	err = tx.EndTx()
	if err != nil {
		storage.l.Error("failed to commit transaction", zap.Error(err))
		return valuemetric.ValueMetric{}, err
	}
	return *valret, nil
}

func (storage *DBStorage) Get(name string) (valuemetric.ValueMetric, error) {
	// select from db
	var valret *valuemetric.ValueMetric

	err := storage.db.SelectValue(name, func(n string, k int, d int64, v float64) error {
		kind, errK := valuemetric.GetKindInt(k)
		if errK != nil {
			return errK
		}
		valret, errK = valuemetric.ConvertToValueMetricInt(kind, &d, &v)
		return errK
	})

	if err != nil {
		return valuemetric.ValueMetric{}, err
	}
	return *valret, nil
}

func (storage *DBStorage) ReadAll(prog memstorage.FuncReadAllMetric) error {
	err := storage.db.SelectValueAll(func(n string, k int, d int64, v float64) error {
		kind, errK := valuemetric.GetKindInt(k)
		if errK != nil {
			return errK
		}
		val, err := valuemetric.ConvertToValueMetricInt(kind, &d, &v)
		if err != nil {
			return err
		}
		return prog(n, *val)
	})
	if err != nil {
		return err
	}
	return nil
}

func (storage *DBStorage) ReadAllClearCounters(prog memstorage.FuncReadAllMetric) error {
	// select all from db and call prog for each record, and clear counters
	return nil
}
