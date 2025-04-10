package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Setint64(t *testing.T) {
	var a int64 = 10
	tests := []struct {
		name string
		val  *int64
		want int64
	}{
		{name: "Test Setint64", val: &a, want: 10},
		{name: "Test Setint64 nil", val: nil, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Setint64(tt.val)
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_Setfloat64(t *testing.T) {
	a := 10.10
	tests := []struct {
		name string
		val  *float64
		want float64
	}{
		{name: "Test Setfloat64", val: &a, want: 10.10},
		{name: "Test Setfloat64 nil", val: nil, want: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Setfloat64(tt.val)
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_probeDefault(t *testing.T) {
	t.Run("Test probe_Default", func(t *testing.T) {
		var a error
		assert.Equal(t, probeDefault(a), true)
	})
}

func Test_RetryTimes(t *testing.T) {
	t.Run("Test RetryTimes", func(t *testing.T) {
		assert.Equal(t, RetryTimes(), []int{1000, 3000, 5000})
	})
}
