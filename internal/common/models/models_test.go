package models

import (
	"testing"

	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
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
	} else {
		if tt.Delta != nil {
			return false
		}
	}
	if testModel.Value != nil {
		if tt.Value == nil {
			return false
		}
		if *tt.Value != *testModel.Value {
			return false
		}
	} else {
		if tt.Value != nil {
			return false
		}
	}
	return true
}

func Test_ConvertMetricToModel(t *testing.T) {
	//func (valModels *Metrics) ConvertMetricToModel(name string, valMetrics valuemetric.ValueMetric)
	valInt := valuemetric.ConvertToIntValueMetric(44)
	valFloat := valuemetric.ConvertToFloatValueMetric(44.66)

	tests := []struct {
		name      string
		valueName string
		value     *valuemetric.ValueMetric
		want      Metrics
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
	//func (valModels *Metrics) ConvertMetricToValue() string
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
