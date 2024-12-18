package models

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/4aleksei/metricscum/internal/common/repository"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (valModels *Metrics) ConvertMetricToModel(name string, valMetrics repository.ValueMetric) {
	valModels.ID = name
	valModels.MType = valMetrics.GetTypeStr()
	valModels.Delta = valMetrics.ValueInt()
	valModels.Value = valMetrics.ValueFloat()
}

func (valModels *Metrics) JSONDecode(body io.ReadCloser) error {
	dec := json.NewDecoder(body)
	err := dec.Decode(valModels)
	return err
}

func (valModels *Metrics) JSONEncode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return valModels.JSONEncodeBytes(buf)
}

func (valModels *Metrics) JSONEncodeBytes(buf *bytes.Buffer) (*bytes.Buffer, error) {
	enc := json.NewEncoder(buf)
	err := enc.Encode(valModels)
	return buf, err
}
