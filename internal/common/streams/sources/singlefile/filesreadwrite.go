package singlefile

import (
	"bufio"
	"io"
	"os"
)

type (
	producer struct {
		file   *os.File
		writer *bufio.Writer
	}
	consumer struct {
		file   *os.File
		reader *bufio.Reader
	}

	fileReader struct {
		filename string
		reader   *consumer
	}

	fileWriter struct {
		filename string
		writer   *producer
	}
)

const (
	defaultMode os.FileMode = 0666
)

func NewReader(filename string) *fileReader {
	return &fileReader{
		filename: filename,
	}
}

func NewWriter(filename string) *fileWriter {
	return &fileWriter{
		filename: filename,
	}
}

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

func (filestor *fileReader) OpenReader() (io.Reader, error) {
	var err error
	filestor.reader, err = newConsumer(filestor.filename)
	if err != nil {
		return nil, err
	}
	return filestor.reader.reader, nil
}

func (filestor *fileReader) CloseRead() error {
	if filestor.reader != nil {
		defer func() { filestor.reader = nil }()
		if err := filestor.reader.file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (filestor *fileWriter) OpenWriter() (io.Writer, error) {
	var err error
	filestor.writer, err = newProducer(filestor.filename)
	if err != nil {
		return nil, err
	}
	return filestor.writer.writer, nil
}

func (filestor *fileWriter) CloseWrite() error {
	if filestor.writer != nil {
		defer func() { filestor.writer = nil }()
		if err := filestor.writer.writer.Flush(); err != nil {
			filestor.writer.file.Close()
			return err
		}
		if err := filestor.writer.file.Close(); err != nil {
			return err
		}
	}
	return nil
}
