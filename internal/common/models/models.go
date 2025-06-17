// Package models
package models

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

// Metrics - Import/Export to json
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

func (valModels *Metrics) ConvertMetricToModel(name string, valMetrics valuemetric.ValueMetric) {
	valModels.ID = name
	valModels.MType = valMetrics.GetTypeStr()
	valModels.Delta = valMetrics.ValueInt()
	valModels.Value = valMetrics.ValueFloat()
}

func (valModels *Metrics) ConvertMetricToValue() string {
	if valModels.Delta != nil {
		return strconv.FormatInt(*valModels.Delta, 10)
	} else if valModels.Value != nil {
		return strconv.FormatFloat(*valModels.Value, 'f', -1, 64)
	}
	return ""
}

func (valModels *Metrics) JSONDecode(body io.ReadCloser) error {
	dec := json.NewDecoder(body)
	err := dec.Decode(valModels)
	return err
}

func (valModels *Metrics) JSONEncodeBytes(w io.Writer) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(valModels)
	return err
}

func JSONSDecode(body io.ReadCloser) ([]Metrics, error) {
	var valModels []Metrics
	dec := json.NewDecoder(body)
	err := dec.Decode(&valModels)
	return valModels, err
}

func JSONSEncodeBytes(w io.Writer, val []Metrics) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(val)
	return err
}
