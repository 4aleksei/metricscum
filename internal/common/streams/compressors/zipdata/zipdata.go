// Package zipdata
package zipdata

import (
	"compress/gzip"
	"io"
)

type (
	gzipDecWr struct {
		gzipCompress *gzip.Writer
	}

	gzipDecRd struct {
		gzipDecompress *gzip.Reader
	}
)

func NewReader() *gzipDecRd {
	return &gzipDecRd{}
}

func NewWriter() *gzipDecWr {
	return &gzipDecWr{}
}

func (gzipencdec *gzipDecRd) OpenReader(r io.Reader) (io.Reader, error) {
	var err error
	gzipencdec.gzipDecompress, err = gzip.NewReader(r)
	return gzipencdec.gzipDecompress, err
}

func (gzipencdec *gzipDecWr) OpenWriter(w io.Writer) (io.Writer, error) {
	gzipencdec.gzipCompress = gzip.NewWriter(w)
	return gzipencdec.gzipCompress, nil
}

func (gzipencdec *gzipDecRd) CloseRead() error {
	if gzipencdec.gzipDecompress != nil {
		defer func() { gzipencdec.gzipDecompress = nil }()
		err := gzipencdec.gzipDecompress.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (gzipencdec *gzipDecWr) CloseWrite() error {
	if gzipencdec.gzipCompress != nil {
		defer func() { gzipencdec.gzipCompress = nil }()
		if err := gzipencdec.gzipCompress.Flush(); err != nil {
			gzipencdec.gzipCompress.Close()
			return err
		}
		if err := gzipencdec.gzipCompress.Close(); err != nil {
			return err
		}
	}
	return nil
}
