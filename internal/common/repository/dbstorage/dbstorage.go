package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/store"
	"github.com/4aleksei/metricscum/internal/common/utils"
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

const defaultTimeoutPing int = 500

func (storage *DBStorage) PingContext(ctx context.Context) error {
	err := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		ctxP, cancel := context.WithTimeout(ctx, time.Duration(defaultTimeoutPing)*time.Millisecond)
		defer cancel()
		return storage.db.DB.PingContext(ctxP)
	}, store.ProbePG)
	return err
}

func (storage *DBStorage) AddMulti(ctx context.Context, modval []models.Metrics) (*[]models.Metrics, error) {
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

func (storage *DBStorage) Add(ctx context.Context, name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
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

func (storage *DBStorage) Get(ctx context.Context, name string) (valuemetric.ValueMetric, error) {
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

func (storage *DBStorage) ReadAll(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	errR := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		err := storage.db.SelectValueAll(ctx, func(n string, k int, d int64, v float64) error {
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
		return err
	}, store.ProbePG)

	return errR
}

func (storage *DBStorage) ReadAllClearCounters(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	return nil
}
