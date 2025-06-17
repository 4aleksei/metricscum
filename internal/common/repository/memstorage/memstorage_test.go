package memstorage

import (
	"context"
	"errors"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/stretchr/testify/assert"
)

func Test_PingContext(t *testing.T) {
	n := NewStore()
	tests := []struct {
		want error
		name string
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
	n := NewStore()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vFF := valuemetric.ConvertToFloatValueMetric(77.77)
	vI := valuemetric.ConvertToIntValueMetric(55)
	vII := valuemetric.ConvertToIntValueMetric(55 + 55)
	tests := []struct {
		wantErr error
		name    string
		valName string
		val     valuemetric.ValueMetric
		wantVal valuemetric.ValueMetric
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
	n := NewStore()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	_, _ = n.Add(context.Background(), nameOfTest1, *vI)
	_, _ = n.Add(context.Background(), nameOfTest2, *vF)

	tests := []struct {
		wantErr error
		name    string
		valName string
		wantVal valuemetric.ValueMetric
	}{
		{name: "Test Get Int", valName: nameOfTest1, wantVal: *vI, wantErr: nil},
		{name: "Test Get Float", valName: nameOfTest2, wantVal: *vF, wantErr: nil},
		{name: "Test Get NotFound", valName: "testNotFound", wantVal: *vF, wantErr: ErrNotFoundName},
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
	n := NewStore()
	var a int64 = 100
	b := 100.1
	modval := make([]models.Metrics, 0)
	modval = append(modval, models.Metrics{ID: "Test1", MType: "counter", Delta: &a})
	modval = append(modval, models.Metrics{ID: "Test2", MType: "gauge", Value: &b})

	modval2 := make([]models.Metrics, 0)
	modval2 = append(modval2, models.Metrics{ID: "Test3", MType: "Badcounter", Delta: &a})

	modval3 := make([]models.Metrics, 0)
	modval3 = append(modval3, models.Metrics{ID: "", MType: "counter", Delta: &a})

	modval4 := make([]models.Metrics, 0)
	modval4 = append(modval4, models.Metrics{ID: "Test4", MType: "counter"})

	tests := []struct {
		wantErr error
		name    string
		val     []models.Metrics
		wantVal []models.Metrics
	}{
		{name: "Test AddMulti", val: modval, wantVal: modval, wantErr: nil},
		{name: "Test AddMulti Error Kind", val: modval2, wantVal: modval2, wantErr: valuemetric.ErrBadTypeValue},
		{name: "Test AddMulti Error Name", val: modval3, wantVal: modval3, wantErr: ErrBadName},
		{name: "Test AddMulti Error Val", val: modval4, wantVal: modval4, wantErr: valuemetric.ErrBadValue},
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
	n := NewStore()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	_, _ = n.Add(context.Background(), nameOfTest1, *vI)
	_, _ = n.Add(context.Background(), nameOfTest2, *vF)

	tests := []struct {
		wantErr error
		name    string
		valName string
		wantVal valuemetric.ValueMetric
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
	n := NewStore()
	vF := valuemetric.ConvertToFloatValueMetric(55.55)
	vI := valuemetric.ConvertToIntValueMetric(55)
	nameOfTest1 := "Test1"
	nameOfTest2 := "Test2"
	_, _ = n.Add(context.Background(), nameOfTest1, *vI)
	_, _ = n.Add(context.Background(), nameOfTest2, *vF)

	tests := []struct {
		wantErr error
		name    string
		valName string
		wantVal valuemetric.ValueMetric
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
