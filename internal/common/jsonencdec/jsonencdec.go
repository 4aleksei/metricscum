package jsonencdec

import (
	"encoding/json"
	"io"

	"github.com/4aleksei/metricscum/internal/common/models"
)

type jsonencDec struct {
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewJSONEncDec() *jsonencDec {
	return &jsonencDec{}
}

func (jsonencdec *jsonencDec) OpenReader(r io.Reader) {
	jsonencdec.decoder = json.NewDecoder(r)
}

func (jsonencdec *jsonencDec) OpenWriter(w io.Writer) {
	jsonencdec.encoder = json.NewEncoder(w)
}

func (jsonencdec *jsonencDec) WriteData(d *models.Metrics) error {
	return jsonencdec.encoder.Encode(d)
}

func (jsonencdec *jsonencDec) ReadData(d *models.Metrics) error {
	return jsonencdec.decoder.Decode(d)
}

func (jsonencdec *jsonencDec) Close() {
	if jsonencdec.encoder != nil {
		jsonencdec.encoder = nil
	}
	if jsonencdec.decoder != nil {
		jsonencdec.decoder = nil
	}
}
