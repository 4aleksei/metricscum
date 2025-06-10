package service

import (
	"testing"

	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
)

func Test_CheckType(t *testing.T) {
	h := &HandlerStore{}
	tests := []struct {
		want  error
		name  string
		value string
	}{
		{name: "empty Tyoe", value: "", want: valuemetric.ErrBadTypeValue},
		{name: "counter Type", value: "counter", want: nil},
		{name: "gauge Type", value: "gauge", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := h.CheckType(tt.value); got != tt.want {
				t.Errorf("CheckType(string) = %v, want %v", got, tt.want)
			}
		})
	}
}
