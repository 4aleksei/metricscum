package longtermfile

import (
	"io"

	"github.com/4aleksei/metricscum/internal/common/models"
)

type (
	sourceReader interface {
		OpenReader() (io.Reader, error)
		CloseRead() error
	}

	middleReaders interface {
		OpenReader(io.Reader) (io.Reader, error)
		CloseRead() error
	}

	sourceWriter interface {
		OpenWriter() (io.Writer, error)
		CloseWrite() error
	}

	middleWriters interface {
		OpenWriter(io.Writer) (io.Writer, error)
		CloseWrite() error
	}

	modelsReader interface {
		OpenReader(io.Reader)
		ReadData(*models.Metrics) error
		CloseRead()
	}

	modelsWriter interface {
		OpenWriter(io.Writer)
		WriteData(*models.Metrics) error
		CloseWrite()
	}

	longtermfile struct {
		sourcesRead   sourceReader
		modelsdecoder modelsReader
		sourcesWrite  sourceWriter
		modelsencoder modelsWriter
		middleReads   []middleReaders
		middleWriters []middleWriters
	}
)

func NewLongTerm(r sourceReader, mr modelsReader, w sourceWriter, mw modelsWriter) *longtermfile {
	l := new(longtermfile)
	l.sourcesRead = r
	l.sourcesWrite = w
	l.modelsdecoder = mr
	l.modelsencoder = mw
	return l
}

func (l *longtermfile) UseForReader(m middleReaders) {
	l.middleReads = append(l.middleReads, m)
}
func (l *longtermfile) UseForWriter(m middleWriters) {
	l.middleWriters = append(l.middleWriters, m)
}

func (l *longtermfile) OpenReader() error {
	readIo, err := l.sourcesRead.OpenReader()
	if err != nil {
		return err
	}

	for _, middle := range l.middleReads {
		readIo, err = middle.OpenReader(readIo)
		if err != nil {
			_ = l.CloseRead()
			return err
		}
	}
	l.modelsdecoder.OpenReader(readIo)
	return nil
}

func (l *longtermfile) OpenWriter() error {
	writeIo, err := l.sourcesWrite.OpenWriter()
	if err != nil {
		return err
	}

	for _, middle := range l.middleWriters {
		writeIo, err = middle.OpenWriter(writeIo)
		if err != nil {
			_ = l.CloseWrite()
			return err
		}
	}
	l.modelsencoder.OpenWriter(writeIo)
	return nil
}

func (l *longtermfile) WriteData(m *models.Metrics) error {
	return l.modelsencoder.WriteData(m)
}

func (l *longtermfile) ReadData(m *models.Metrics) error {
	return l.modelsdecoder.ReadData(m)
}

func (l *longtermfile) CloseRead() error {
	err := l.sourcesRead.CloseRead()
	if err != nil {
		return err
	}
	for _, middle := range l.middleReads {
		err = middle.CloseRead()
		if err != nil {
			return err
		}
	}
	l.modelsdecoder.CloseRead()
	return err
}

func (l *longtermfile) CloseWrite() error {
	l.modelsencoder.CloseWrite()
	var err error
	for _, middle := range l.middleWriters {
		err = middle.CloseWrite()
		if err != nil {
			return err
		}
	}

	return l.sourcesWrite.CloseWrite()
}
