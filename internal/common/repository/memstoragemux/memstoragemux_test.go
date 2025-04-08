package memstoragemux

import (
	"context"
	"errors"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

func Test_Add(t *testing.T) {
	// Add(ctx context.Context, name string, val valuemetric.ValueMetric) (valuemetric.ValueMetric, error)

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
					t.Errorf("ConvertToValueMetricInt = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if got != tt.wantVal {
					t.Errorf("ConvertToValueMetricInt = %v, want_val %v ", got, tt.wantVal)
				}
			}
		})
	}
}
