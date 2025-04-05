// Package dbstorage
package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"database/sql"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/store"
	"github.com/4aleksei/metricscum/internal/common/store/pg"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"go.uber.org/zap"
)

type DBStorage struct {
	db store.Store
	l  *zap.Logger
}

var (
	ErrBadValue = errors.New("invalid value")
	ErrBadName  = errors.New("no name")
	ErrNoDB     = errors.New("no db")
)

func NewStoreDB(db *pg.DB, l *zap.Logger) *DBStorage {
	return &DBStorage{db: db,
		l: l}
}

const defaultTimeoutPing int = 500

func (storage *DBStorage) PingContext(ctx context.Context) error {
	err := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		ctxP, cancel := context.WithTimeout(ctx, time.Duration(defaultTimeoutPing)*time.Millisecond)
		defer cancel()
		return storage.db.Ping(ctxP)
	}, pg.ProbePG)
	return err
}

const limitbatch int = 5

func (storage *DBStorage) AddMulti(ctx context.Context, modval []models.Metrics) ([]models.Metrics, error) {
	resmodels := make([]models.Metrics, 0, len(modval))
	sm := make([]store.Metrics, len(modval))
	var i = 0
	for _, valModel := range modval {
		kind, errKind := valuemetric.GetKind(valModel.MType)
		if errKind != nil {
			return nil, fmt.Errorf("failed kind %w", errKind)
		}
		if valModel.ID == "" {
			return nil, fmt.Errorf("failed %w", ErrBadName)
		}

		sm[i].Kind = int(kind)
		sm[i].Name = valModel.ID
		sm[i].Delta = sql.NullInt64{Valid: valModel.Delta != nil, Int64: utils.Setint64(valModel.Delta)}
		sm[i].Value = sql.NullFloat64{Valid: valModel.Value != nil, Float64: utils.Setfloat64(valModel.Value)}

		i++
	}

	var valret *valuemetric.ValueMetric

	err := storage.db.Upserts(ctx, sm, limitbatch, func(n string, k int, d int64, v float64) error {
		kind, errK := valuemetric.GetKindInt(k)
		if errK != nil {
			return errK
		}
		valret, errK = valuemetric.ConvertToValueMetricInt(kind, &d, &v)
		var valNewModel models.Metrics
		valNewModel.ConvertMetricToModel(n, *valret)
		resmodels = append(resmodels, valNewModel)
		return errK
	})

	if err != nil {
		storage.l.Error("failed to upserts transaction", zap.Error(err))
		return nil, err
	}

	return resmodels, nil
}

func (storage *DBStorage) Add(ctx context.Context, name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error) {
	var modval store.Metrics
	modval.Kind = val.GetKind()
	modval.Name = name
	modval.Delta = sql.NullInt64{Valid: val.ValueInt() != nil, Int64: utils.Setint64(val.ValueInt())}
	modval.Value = sql.NullFloat64{Valid: val.ValueFloat() != nil, Float64: utils.Setfloat64(val.ValueFloat())}

	var valret *valuemetric.ValueMetric
	err := storage.db.Upsert(ctx, modval, func(n string, k int, d int64, v float64) error {
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

	return *valret, nil
}

func (storage *DBStorage) Get(ctx context.Context, name string) (valuemetric.ValueMetric, error) {
	var valret *valuemetric.ValueMetric

	err := storage.db.SelectValue(ctx, name, func(n string, k int, d int64, v float64) error {
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
	}, pg.ProbePG)

	return errR
}

func (storage *DBStorage) ReadAllClearCounters(ctx context.Context, prog memstorage.FuncReadAllMetric) error {
	return nil
}
