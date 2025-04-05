package valuemetric

import (
	"errors"
	"fmt"
	"testing"
)

func BenchmarkConvertValueMetricToPlain(b *testing.B) {
	var val = ValueMetric{
		kind:     kindInt64,
		valueInt: 110,
	}
	b.Run("old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ConvertValueMetricToPlain(val)
		}
	})
	b.Run("optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ConvertValueMetricToPlainOpt(val)
		}
	})
}

func BenchmarkSprintf(b *testing.B) {
	key := "key"
	value := "343324234343"
	triesN := 60000

	b.Run("old", func(b *testing.B) {
		var valstr string
		for i := 0; i < triesN; i++ {
			valstr += fmt.Sprintf("%s : %s\n", key, value)
		}
	})
	b.Run("optimized", func(b *testing.B) {
		var valstr string
		for i := 0; i < triesN; i++ {
			valstr += key + " : " + value + "\n"
		}
	})
}

func Test_ConvertToValueMetricInt(t *testing.T) {
	// ConvertToValueMetricInt(kind valueKind, delta *int64, value *float64) (*ValueMetric, error)
	var testInt int64 = 44
	testFloat := 66.33

	tests := []struct {
		name    string
		kind    valueKind
		delta   *int64
		value   *float64
		wantVal ValueMetric
		wantErr error
	}{
		{name: "Test kindInt64", kind: kindInt64, delta: &testInt, value: nil, wantVal: ValueMetric{kindInt64, 44, 0}, wantErr: nil},
		{name: "Test kindFloat64", kind: kindFloat64, delta: nil, value: &testFloat, wantVal: ValueMetric{kindFloat64, 0, 66.33}, wantErr: nil},
		{name: "Test Bad kindInt64", kind: kindInt64, delta: nil, value: nil, wantVal: ValueMetric{kindInt64, 44, 0}, wantErr: ErrBadValue},
		{name: "Test Bad kindFloat64", kind: kindFloat64, delta: nil, value: nil, wantVal: ValueMetric{kindFloat64, 44, 0}, wantErr: ErrBadValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ConvertToValueMetricInt(tt.kind, tt.delta, tt.value)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ConvertToValueMetricInt = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if *got != tt.wantVal {
					t.Errorf("ConvertToValueMetricInt = %v, want_val %v ", *got, tt.wantVal)
				}
			}
		})
	}
}

func Test_ConvertToValueMetric(t *testing.T) {
	//ConvertToValueMetric(kind valueKind, valstr string) (*ValueMetric, error)
	tests := []struct {
		name    string
		kind    valueKind
		value   string
		wantVal ValueMetric
		wantErr error
	}{
		{name: "Test kindInt64", kind: kindInt64, value: "44", wantVal: ValueMetric{kindInt64, 44, 0}, wantErr: nil},
		{name: "Test kindFloat64", kind: kindFloat64, value: "66.33", wantVal: ValueMetric{kindFloat64, 0, 66.33}, wantErr: nil},
		{name: "Test Bad kindInt64", kind: kindInt64, value: "TestErre44", wantVal: ValueMetric{kindInt64, 44, 0}, wantErr: ErrBadValue},
		{name: "Test Bad kindFloat64", kind: kindFloat64, value: "TestErre44.33", wantVal: ValueMetric{kindFloat64, 44, 0}, wantErr: ErrBadValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ConvertToValueMetric(tt.kind, tt.value)
			if gotErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ConvertToValueMetric = %v, want_err %v ", gotErr, tt.wantErr)
				}
			} else {
				if *got != tt.wantVal {
					t.Errorf("ConvertToValueMetric = %v, want_val %v ", *got, tt.wantVal)
				}
			}
		})
	}
}

func Test_ConvertValueMetricToPlain(t *testing.T) {
	//func ConvertValueMetricToPlain(val ValueMetric) (a, b string)

	tests := []struct {
		name     string
		value    ValueMetric
		wantValA string
		wantValB string
	}{
		{name: "Test kindInt64", value: ValueMetric{kind: kindInt64, valueInt: 44, valueFloat: 0}, wantValA: "counter", wantValB: "44"},
		{name: "Test kindFloat64", value: ValueMetric{kind: kindFloat64, valueInt: 0, valueFloat: 44.5}, wantValA: "gauge", wantValB: "44.5"},
		{name: "Test kindEmpty", value: ValueMetric{kind: kindBadEmpty, valueInt: 0, valueFloat: 0}, wantValA: defaultNAN, wantValB: defaultNAN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := ConvertValueMetricToPlain(tt.value)
			if gotA != tt.wantValA {
				t.Errorf("ConvertValueMetricToPlain = %v, want_err %v ", gotA, tt.wantValA)
			}

			if gotB != tt.wantValB {
				t.Errorf("ConvertValueMetricToPlain = %v, want_err %v ", gotB, tt.wantValB)
			}
		})
	}
}

func Test_ConvertValueMetricToPlainOpt(t *testing.T) {
	tests := []struct {
		name     string
		value    ValueMetric
		wantValA string
		wantValB string
	}{
		{name: "Test kindInt64", value: ValueMetric{kind: kindInt64, valueInt: 44, valueFloat: 0}, wantValA: "44"},
		{name: "Test kindFloat64", value: ValueMetric{kind: kindFloat64, valueInt: 0, valueFloat: 44.5}, wantValA: "44.5"},
		{name: "Test kindEmpty", value: ValueMetric{kind: kindBadEmpty, valueInt: 0, valueFloat: 0}, wantValA: defaultNAN, wantValB: defaultNAN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA := ConvertValueMetricToPlainOpt(tt.value)
			if gotA != tt.wantValA {
				t.Errorf("ConvertValueMetricToPlainOpt = %v, want_err %v ", gotA, tt.wantValA)
			}
		})
	}
}

func Test_KindOf(t *testing.T) {
	tests := []struct {
		name  string
		value valueKind
		val   ValueMetric
		want  bool
	}{
		{name: "Test kindInt64", value: kindInt64, val: ValueMetric{kind: kindInt64, valueInt: 44, valueFloat: 0}, want: true},
		{name: "Test kindFloat64", value: kindFloat64, val: ValueMetric{kind: kindFloat64, valueInt: 0, valueFloat: 44.5}, want: true},
		{name: "Test kindEmpty", value: kindBadEmpty, val: ValueMetric{kind: kindBadEmpty, valueInt: 0, valueFloat: 0}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.KindOf(tt.value)
			if got != tt.want {
				t.Errorf("KindOf = %v, valueMetric=%v, want %v ", tt.value, tt.val, tt.want)
			}
		})
	}
}

func Test_ConvertToFloatValueMetric(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  ValueMetric
	}{
		{name: "Test kindFloat64", value: 55.6, want: ValueMetric{kind: kindFloat64, valueInt: 0, valueFloat: 55.6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToFloatValueMetric(tt.value)
			if *got != tt.want {
				t.Errorf("ConvertToFloatValueMetric = %v, want %v ", tt.value, tt.want)
			}
		})
	}
}

func Test_ConvertToIntValueMetric(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  ValueMetric
	}{
		{name: "Test kindFloat64", value: 55, want: ValueMetric{kind: kindInt64, valueInt: 55, valueFloat: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToIntValueMetric(tt.value)
			if *got != tt.want {
				t.Errorf("ConvertToIntValueMetric = %v, want %v ", tt.value, tt.want)
			}
		})
	}
}

func Test_GetTypeStr(t *testing.T) {
	tests := []struct {
		name  string
		value ValueMetric
		want  string
	}{
		{name: "Test kindInt64", value: ValueMetric{kind: kindInt64, valueInt: 55, valueFloat: 0}, want: "counter"},
		{name: "Test kindFloat64", value: ValueMetric{kind: kindFloat64, valueInt: 0, valueFloat: 55.66}, want: "gauge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.GetTypeStr()
			if got != tt.want {
				t.Errorf("GetTypeStr = %v, want %v ", tt.value, tt.want)
			}
		})
	}
}
