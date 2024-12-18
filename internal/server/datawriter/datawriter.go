package datawriter

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/server/config"
	"go.uber.org/zap"
)

type ServerDataWriter interface {
	ReadAll(repository.FuncReadAllMetric) error
	Add(string, repository.ValueMetric) repository.ValueMetric
}

type DataWriterHandler struct {
	store ServerDataWriter
	cfg   *config.Config
}

func NewAppWriter(store ServerDataWriter, cfg *config.Config) *DataWriterHandler {
	app := new(DataWriterHandler)
	app.store = store
	app.cfg = cfg
	return app
}

func (app *DataWriterHandler) RunRutine() {
	if app.cfg.WriteInterval > 0 {
		go app.mainWriter()
	}
}

func (app *DataWriterHandler) mainWriter() {

	for {

		time.Sleep(time.Duration(app.cfg.WriteInterval) * time.Second)
		app.DoWriteData()

	}

}

type producer struct {
	file   *os.File
	writer *bufio.Writer
}

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (app *DataWriterHandler) DoWriteData() {

	fileData, err := newProducer(app.cfg.FilePath)
	if err != nil {
		logger.Log.Debug("error writing response", zap.Error(err))
		return
	}
	defer func() { fileData.writer.Flush(); fileData.file.Close() }()
	valNewModel := new(models.Metrics)
	err = app.store.ReadAll(func(key string, val repository.ValueMetric) error {
		valNewModel.ConvertMetricToModel(key, val)

		if errson := valNewModel.JSONEncodeBytes(fileData.writer); errson != nil {
			logger.Log.Debug("error writing json", zap.Error(errson))
			return err
		}
		fileData.writer.WriteByte('\n')
		return nil
	})

	if err != nil {
		return
	}

}

type consumer struct {
	file   *os.File
	reader *bufio.Reader
}

func newConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (app *DataWriterHandler) ReadData() error {

	fileData, err := newConsumer(app.cfg.FilePath)
	if err != nil {
		logger.Log.Debug("error read file", zap.Error(err))
		return err
	}
	defer fileData.file.Close()

	valNewModel := new(models.Metrics)
	dec := json.NewDecoder(fileData.reader)

	for {
		if errson := dec.Decode(valNewModel); errson != nil {
			return errson
		}
		kind, errKind := repository.GetKind(valNewModel.MType)
		if errKind != nil {
			return errKind
		}
		if valNewModel.ID == "" {
			return errors.New("no name")
		}
		val, err := repository.ConvertToValueMetricInt(kind, valNewModel.Delta, valNewModel.Value)
		if err != nil {
			return err
		}
		_ = app.store.Add(valNewModel.ID, *val)
	}

}
