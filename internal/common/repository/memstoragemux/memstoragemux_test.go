package memstoragemux

import (
	"context"
	"errors"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/stretchr/testify/assert"
)

func Test_PingContext(t *testing.T) {
	n := NewStoreMux()
	tests := []struct {
		name string
		want error
	}{
		{name: "Test Ping", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.PingContext(context.Background())
			if got != tt.want {
				t.Errorf("PingContext = %v ", got)
			}
		})
	}
}

func Test_Add(t *testing.T) {
	n := NewStoreMux()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vFF := valuemetric.ConvertToFloatValueMetric(77.77)
	vI := valuemetric.ConvertToIntValueMetric(55)
	vII := valuemetric.ConvertToIntValueMetric(55 + 55)
	tests := []struct {
		name    string
		valName string
		val     valuemetric.ValueMetric
		wantVal valuemetric.ValueMetric
		wantErr error
	}{
		{name: "Test Add Int", valName: "test1", val: *vI, wantVal: *vI, wantErr: nil},
		{name: "Test Add Float", valName: "test2", val: *vF, wantVal: *vF, wantErr: nil},
		{name: "Test Increment Int", valName: "test1", val: *vI, wantVal: *vII, wantErr: nil},
		{name: "Test Replace Float", valName: "test2", val: *vFF, wantVal: *vFF, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := n.Add(context.Background(), tt.valName, tt.val)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("Add = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if got != tt.wantVal {
					t.Errorf("Add = %v, want_val %v ", got, tt.wantVal)
				}
			}
		})
	}
}

func Test_Get(t *testing.T) {
	n := NewStoreMux()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	n.Add(context.Background(), nameOfTest1, *vI)
	n.Add(context.Background(), nameOfTest2, *vF)

	tests := []struct {
		name    string
		valName string
		wantVal valuemetric.ValueMetric
		wantErr error
	}{
		{name: "Test Get Int", valName: nameOfTest1, wantVal: *vI, wantErr: nil},
		{name: "Test Get Float", valName: nameOfTest2, wantVal: *vF, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := n.Get(context.Background(), tt.valName)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("Get = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if got != tt.wantVal {
					t.Errorf("Get = %v, want_val %v ", got, tt.wantVal)
				}
			}
		})
	}
}

func Test_AddMulti(t *testing.T) {
	n := NewStoreMux()
	var a int64 = 100
	b := 100.1
	modval := make([]models.Metrics, 0)
	modval = append(modval, models.Metrics{ID: "Test1", MType: "counter", Delta: &a})
	modval = append(modval, models.Metrics{ID: "Test2", MType: "gauge", Value: &b})

	tests := []struct {
		name    string
		val     []models.Metrics
		wantVal []models.Metrics
		wantErr error
	}{
		{name: "Test AddMulti", val: modval, wantVal: modval, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := n.AddMulti(context.Background(), tt.val)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("AddMulti = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				assert.ElementsMatch(t, got, tt.wantVal)
			}
		})
	}
}

func Test_ReadAllClearCounters(t *testing.T) {
	n := NewStoreMux()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	n.Add(context.Background(), nameOfTest1, *vI)
	n.Add(context.Background(), nameOfTest2, *vF)

	tests := []struct {
		name    string
		valName string
		wantVal valuemetric.ValueMetric
		wantErr error
	}{
		{name: "Test Get Int", valName: nameOfTest1, wantVal: *vI, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := n.ReadAllClearCounters(context.Background(), func(name string, val valuemetric.ValueMetric) error {

				return nil
			})
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ReadAllClearCounter = %v, want_err %v ", gotErr, tt.wantErr)
				}
			}
			got, _ := n.Get(context.Background(), nameOfTest1)
			if got.ValueInt() == nil {
				t.Errorf("ReadAllClearCounter = %v, want_err %v ", gotErr, tt.wantErr)
			}

			if *got.ValueInt() != 0 {
				t.Errorf("ReadAllClearCounter = %v, want_err %v ", gotErr, tt.wantErr)
			}

		})
	}
}

func Test_ReadAll(t *testing.T) {
	n := NewStoreMux()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	n.Add(context.Background(), nameOfTest1, *vI)
	n.Add(context.Background(), nameOfTest2, *vF)

	tests := []struct {
		name    string
		valName string
		wantVal valuemetric.ValueMetric
		wantErr error
	}{
		{name: "Test Get Int", valName: nameOfTest1, wantVal: *vI, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := n.ReadAll(context.Background(), func(name string, val valuemetric.ValueMetric) error {

				return nil
			})
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ReadlAll = %v, want_err %v ", gotErr, tt.wantErr)
				}
			}
		})
	}
}
