// Package jsonencdec
package jsonencdec

import (
	"encoding/json"
	"io"

	"github.com/4aleksei/metricscum/internal/common/models"
)

type (
	jsonencEnc struct {
		encoder *json.Encoder
	}
	jsonencDec struct {
		decoder *json.Decoder
	}
)

func NewReader() *jsonencDec {
	return &jsonencDec{}
}

func NewWriter() *jsonencEnc {
	return &jsonencEnc{}
}

func (jsonencdec *jsonencDec) OpenReader(r io.Reader) {
	jsonencdec.decoder = json.NewDecoder(r)
}

func (jsonencdec *jsonencEnc) OpenWriter(w io.Writer) {
	jsonencdec.encoder = json.NewEncoder(w)
}

func (jsonencdec *jsonencEnc) WriteData(d *models.Metrics) error {
	return jsonencdec.encoder.Encode(d)
}

func (jsonencdec *jsonencDec) ReadData(d *models.Metrics) error {
	return jsonencdec.decoder.Decode(d)
}

func (jsonencdec *jsonencDec) CloseRead() {
	if jsonencdec.decoder != nil {
		jsonencdec.decoder = nil
	}
}

func (jsonencdec *jsonencEnc) CloseWrite() {
	if jsonencdec.encoder != nil {
		jsonencdec.encoder = nil
	}
}
