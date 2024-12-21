package filesreadwrite

import (
	"bufio"
	"io"
	"os"

	"github.com/4aleksei/metricscum/internal/common/models"
)

type producer struct {
	file   *os.File
	writer *bufio.Writer
}
type consumer struct {
	file   *os.File
	reader *bufio.Reader
}

type EncodeDecode interface {
	OpenReader(io.Reader)
	OpenWriter(io.Writer)
	WriteData(*models.Metrics) error
	ReadData(*models.Metrics) error
	Close()
}

type fileStorage struct {
	filename string
	writer   *producer
	reader   *consumer
	encoder  EncodeDecode
}

func NewFileStorage(filename string, encoder EncodeDecode) *fileStorage {
	store := new(fileStorage)
	store.filename = filename
	store.encoder = encoder
	return store
}

const defaultMode os.FileMode = 0666

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultMode)
	if err != nil {
		return nil, err
	}
	return &producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func newConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, defaultMode)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (filestor *fileStorage) WriteData(data *models.Metrics) error {
	return filestor.encoder.WriteData(data)
}
func (filestor *fileStorage) ReadData(data *models.Metrics) error {
	return filestor.encoder.ReadData(data)
}

func (filestor *fileStorage) OpenWriter() error {
	var err error
	filestor.writer, err = newProducer(filestor.filename)
	if err != nil {
		return err
	}
	filestor.encoder.OpenWriter(filestor.writer.writer)
	return nil
}

func (filestor *fileStorage) OpenReader() error {
	var err error
	filestor.reader, err = newConsumer(filestor.filename)
	if err != nil {
		return err
	}
	filestor.encoder.OpenReader(filestor.reader.reader)
	return nil
}

func (filestor *fileStorage) Close() error {
	if filestor.writer != nil {
		defer func() { filestor.writer = nil; filestor.encoder.Close() }()
		if err := filestor.writer.writer.Flush(); err != nil {
			return err
		}
		if err := filestor.writer.file.Close(); err != nil {
			return err
		}
	}
	if filestor.reader != nil {
		defer func() { filestor.reader = nil; filestor.encoder.Close() }()
		if err := filestor.reader.file.Close(); err != nil {
			return err
		}
	}
	return nil
}
