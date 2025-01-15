package resources

import (
	"context"
	"log"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/repository/dbstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/longtermfile"
	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/repository/memstoragemux"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/store"
	"github.com/4aleksei/metricscum/internal/common/streams/compressors/zipdata"
	"github.com/4aleksei/metricscum/internal/common/streams/encoders/jsonencdec"
	"github.com/4aleksei/metricscum/internal/common/streams/sources/singlefile"
	"github.com/4aleksei/metricscum/internal/server/config"
	"go.uber.org/zap"
)

type resoucesMetricsStorage interface {
	Add(string, valuemetric.ValueMetric) (valuemetric.ValueMetric, error)
	Get(string) (valuemetric.ValueMetric, error)
	ReadAll(memstorage.FuncReadAllMetric) error
	PingContext(context.Context) error
	AddMulti([]models.Metrics) (*[]models.Metrics, error)
}

type handleResources struct {
	Store resoucesMetricsStorage
	DB    *store.DB
	FILE  *repository.MemStorageMuxLongTerm
}

func CreateResouces(cfg *config.Config, l *zap.Logger) (*handleResources, error) {
	hs := new(handleResources)
	if cfg.DBcfg.DatabaseDSN != "" {

		db, errDB := store.NewDB(cfg.DBcfg)

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
			storage.DataRun()
			hs.FILE = storage

			hs.Store = storage
		} else {
			hs.Store = memstoragemux.NewStoreMux()
		}
	}
	return hs, nil
}

func (hr *handleResources) Close() error {
	if hr.FILE != nil {
		err := hr.FILE.DataWrite()
		if err != nil {
			log.Println("Error Close File ")
		} else {
			log.Println("File has been closed")
		}
	}

	if hr.DB != nil {
		err := hr.DB.DB.Close()
		if err != nil {
			return err
		} else {
			log.Println("DB has been closed")
		}
	}
	return nil
}
