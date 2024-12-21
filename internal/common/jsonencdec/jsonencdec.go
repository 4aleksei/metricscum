package jsonencdec

import (
	"encoding/json"
	"io"

	"github.com/4aleksei/metricscum/internal/common/models"
)

type JSONEncDec struct {
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewJSONEncDec() *JSONEncDec {
	return &JSONEncDec{}
}

func (jsonencdec *JSONEncDec) OpenReader(body io.Reader) {
	jsonencdec.decoder = json.NewDecoder(body)
}

func (jsonencdec *JSONEncDec) OpenWriter(w io.Writer) {
	jsonencdec.encoder = json.NewEncoder(w)
}

func (jsonencdec *JSONEncDec) WriteData(data *models.Metrics) error {
	return jsonencdec.encoder.Encode(data)
}

func (jsonencdec *JSONEncDec) ReadData(data *models.Metrics) error {
	return jsonencdec.decoder.Decode(data)
}

func (jsonencdec *JSONEncDec) Close() {
	if jsonencdec.encoder != nil {
		jsonencdec.encoder = nil
	}
	if jsonencdec.decoder != nil {
		jsonencdec.decoder = nil
	}
}
