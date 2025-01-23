package resources

import (
	"context"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/repository/dbstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/longtermfile"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/memstoragemux"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/store"
	"github.com/4aleksei/metricscum/internal/common/store/pg"
	"github.com/4aleksei/metricscum/internal/common/streams/compressors/zipdata"
	"github.com/4aleksei/metricscum/internal/common/streams/encoders/jsonencdec"
	"github.com/4aleksei/metricscum/internal/common/streams/sources/singlefile"
	"github.com/4aleksei/metricscum/internal/server/config"
	"go.uber.org/zap"
)

type resoucesMetricsStorage interface {
	Add(context.Context, string, valuemetric.ValueMetric) (valuemetric.ValueMetric, error)
	Get(context.Context, string) (valuemetric.ValueMetric, error)
	ReadAll(context.Context, memstorage.FuncReadAllMetric) error
	PingContext(context.Context) error
	AddMulti(context.Context, []models.Metrics) ([]models.Metrics, error)
}

type handleResources struct {
	Store resoucesMetricsStorage
	DB    store.Store
	FILE  *repository.MemStorageMuxLongTerm
}

func CreateResouces(cfg *config.Config, l *zap.Logger) (*handleResources, error) {
	hs := new(handleResources)
	if cfg.DBcfg.DatabaseDSN != "" {
		db, errDB := pg.NewDB(cfg.DBcfg)

		if errDB != nil {
			l.Debug("DB error", zap.Error(errDB))
			return nil, errDB
		}
		hs.Store = dbstorage.NewStoreDB(db, l)
		hs.DB = db
	} else {
		if cfg.FilePath != "" {
			fileWork := longtermfile.NewLongTerm(singlefile.NewReader(cfg.FilePath),
				jsonencdec.NewReader(), singlefile.NewWriter(cfg.FilePath), jsonencdec.NewWriter())

			fileWork.UseForWriter(zipdata.NewWriter())
			fileWork.UseForReader(zipdata.NewReader())

			storage := repository.NewStoreMuxFiles(&cfg.Repcfg, l, fileWork)
			storage.DataRun(context.TODO())
			hs.FILE = storage

			hs.Store = storage
		} else {
			hs.Store = memstoragemux.NewStoreMux()
		}
	}
	return hs, nil
}

func (hr *handleResources) Close(ctx context.Context) error {
	if hr.FILE != nil {
		err := hr.FILE.DataWrite(ctx)
		if err != nil {
			return err
		}
	}

	if hr.DB != nil {
		hr.DB.Close(ctx)
	}
	return nil
}
