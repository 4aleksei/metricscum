package models

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/stretchr/testify/assert"
)

func checkModels(testModel Metrics, tt Metrics) bool {
	if testModel.ID != tt.ID {
		return false
	}
	if testModel.MType != tt.MType {
		return false
	}
	if testModel.Delta != nil {
		if tt.Delta == nil {
			return false
		}
		if *tt.Delta != *testModel.Delta {
			return false
		}
	} else if tt.Delta != nil {
		return false
	}

	if testModel.Value != nil {
		if tt.Value == nil {
			return false
		}
		if *tt.Value != *testModel.Value {
			return false
		}
	} else if tt.Value != nil {
		return false
	}
	return true
}

func Test_ConvertMetricToModel(t *testing.T) {
	valInt := valuemetric.ConvertToIntValueMetric(44)
	valFloat := valuemetric.ConvertToFloatValueMetric(44.66)

	tests := []struct {
		want      Metrics
		value     *valuemetric.ValueMetric
		name      string
		valueName string
	}{
		{name: "Test Convert to Models Int", valueName: "TestMetr", value: valInt, want: Metrics{ID: "TestMetr", MType: "counter", Delta: valInt.ValueInt()}},
		{name: "Test Convert to Models Float", valueName: "TestMet2", value: valFloat, want: Metrics{ID: "TestMet2", MType: "gauge", Value: valFloat.ValueFloat()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testModel Metrics
			testModel.ConvertMetricToModel(tt.valueName, *tt.value)
			if !checkModels(testModel, tt.want) {
				t.Errorf("ConvertMetricToModel = %v, want %v ", testModel, tt.want)
			}
		})
	}
}

func Test_ConvertMetricToValue(t *testing.T) {
	valInt := valuemetric.ConvertToIntValueMetric(44)
	valFloat := valuemetric.ConvertToFloatValueMetric(44.66)

	tests := []struct {
		name  string
		value Metrics
		want  string
	}{
		{name: "Test Convert to Metrics Int", value: Metrics{ID: "TestMetr", MType: "counter", Delta: valInt.ValueInt()}, want: "44"},
		{name: "Test Convert to Metrics Float", value: Metrics{ID: "TestMet2", MType: "gauge", Value: valFloat.ValueFloat()}, want: "44.66"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.ConvertMetricToValue()
			if got != tt.want {
				t.Errorf("ConvertMetricToValue = %v, want %v ", tt.value, tt.want)
			}
		})
	}
}

func Test_JSONDecode(t *testing.T) {
	valInt := valuemetric.ConvertToIntValueMetric(44)
	valFloat := valuemetric.ConvertToFloatValueMetric(44.66)

	tests := []struct {
		want    Metrics
		wantErr error
		name    string
		value   string
	}{
		{name: "Test Convert Json to Metrics Int", value: "{ \"id\":\"TestMetr\" , \"type\":\"counter\" , \"delta\":44  }", want: Metrics{ID: "TestMetr", MType: "counter", Delta: valInt.ValueInt()}, wantErr: nil},
		{name: "Test Convert Json to Metrics Float", value: "{ \"id\":\"TestMet2\" , \"type\":\"gauge\" , \"value\":44.66  }", want: Metrics{ID: "TestMet2", MType: "gauge", Value: valFloat.ValueFloat()}, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var val Metrics
			body := io.NopCloser(strings.NewReader(tt.value))
			gotErr := val.JSONDecode(body)
			if gotErr != nil {
				if tt.wantErr == nil {
					t.Errorf("JSONDecode = %v, gotErr = %v wantErr %v  ", tt.value, gotErr, tt.wantErr)
				}
			} else {
				assert.Equal(t, val, tt.want)
			}
		})
	}
}

func Test_JSONEncodeBytes(t *testing.T) {
	valInt := valuemetric.ConvertToIntValueMetric(44)
	valFloat := valuemetric.ConvertToFloatValueMetric(44.66)

	tests := []struct {
		value   Metrics
		wantErr error
		name    string
		want    string
	}{
		{name: "Test Convert Metrics to Json  Int", want: "{ \"id\":\"TestMetr\" , \"type\":\"counter\" , \"delta\":44  }", value: Metrics{ID: "TestMetr", MType: "counter", Delta: valInt.ValueInt()}, wantErr: nil},
		{name: "Test Convert Metrics to Json  Float", want: "{ \"id\":\"TestMet2\" , \"type\":\"gauge\" , \"value\":44.66  }", value: Metrics{ID: "TestMet2", MType: "gauge", Value: valFloat.ValueFloat()}, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			gotErr := tt.value.JSONEncodeBytes(io.Writer(&body))

			if gotErr != nil {
				if tt.wantErr == nil {
					t.Errorf("JSONEncodeBytes = %v, gotErr = %v wantErr %v  ", tt.value, gotErr, tt.wantErr)
				}
			} else {
				assert.JSONEq(t, tt.want, body.String())
			}
		})
	}
}

func Test_JSONDecodeSlice(t *testing.T) {
	var a int64 = 44
	b := 44.66
	modval := make([]Metrics, 0)
	modval = append(modval, Metrics{ID: "TestMetr", MType: "counter", Delta: &a})
	modval = append(modval, Metrics{ID: "TestMet2", MType: "gauge", Value: &b})

	tests := []struct {
		wantErr error
		name    string
		value   string
		want    []Metrics
	}{
		{name: "Test Convert Json to Metrics Int", value: "[ { \"id\":\"TestMetr\" , \"type\":\"counter\" , \"delta\":44  } , { \"id\":\"TestMet2\" , \"type\":\"gauge\" , \"value\":44.66  }]", want: modval, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(tt.value))
			got, gotErr := JSONSDecode(body)
			if gotErr != nil {
				if tt.wantErr == nil {
					t.Errorf("JSONDecode = %v, gotErr = %v wantErr %v  ", tt.value, gotErr, tt.wantErr)
				}
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func Test_JSONEncodeBytesSlice(t *testing.T) {
	var a int64 = 44
	b := 44.66
	modval := make([]Metrics, 0)
	modval = append(modval, Metrics{ID: "TestMetr", MType: "counter", Delta: &a})
	modval = append(modval, Metrics{ID: "TestMet2", MType: "gauge", Value: &b})

	tests := []struct {
		wantErr error
		name    string
		want    string
		value   []Metrics
	}{
		{name: "Test Convert Metrics to Json  Int", want: "[{ \"id\":\"TestMetr\" , \"type\":\"counter\" , \"delta\":44  } , { \"id\":\"TestMet2\" , \"type\":\"gauge\" , \"value\":44.66  }]", value: modval, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			gotErr := JSONSEncodeBytes(io.Writer(&body), tt.value)

			if gotErr != nil {
				if tt.wantErr == nil {
					t.Errorf("JSONEncodeBytes = %v, gotErr = %v wantErr %v  ", tt.value, gotErr, tt.wantErr)
				}
			} else {
				assert.JSONEq(t, tt.want, body.String())
			}
		})
	}
}
