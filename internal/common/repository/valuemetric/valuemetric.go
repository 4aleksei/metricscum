package valuemetric

import (
	"errors"
	"fmt"
	"strconv"
)

type valueKind int

const (
	kindBadEmpty valueKind = iota
	kindInt64
	kindFloat64
)

type ValueMetric struct {
	kind       valueKind
	valueInt   int64
	valueFloat float64
}

func (v *ValueMetric) GetTypeStr() string {
	return GetKindStr(v.kind)
}

func (v *ValueMetric) GetKind() int {
	switch v.kind {
	case kindFloat64:
		return int(kindFloat64)
	case kindInt64:
		return int(kindInt64)
	default:
		return int(kindBadEmpty)
	}
}

func (v *ValueMetric) ValueInt() *int64 {
	if v.kind == kindInt64 {
		return &v.valueInt
	}
	return nil
}

func (v *ValueMetric) ValueFloat() *float64 {
	if v.kind == kindFloat64 {
		return &v.valueFloat
	}
	return nil
}

var (
	ErrBadTypeValue = errors.New("invalid typeValue")
	ErrBadValue     = errors.New("error value conversion")
	ErrBadKindType  = errors.New("error kind type")
)

func GetKindInt(k int) (valueKind, error) {
	switch k {
	case int(kindFloat64):
		return kindFloat64, nil
	case int(kindInt64):
		return kindInt64, nil
	default:
		return kindBadEmpty, ErrBadTypeValue
	}
}

func GetKind(typeValue string) (valueKind, error) {
	switch typeValue {
	case "gauge":
		return kindFloat64, nil
	case "counter":
		return kindInt64, nil
	default:
		return kindBadEmpty, ErrBadTypeValue
	}
}

func GetKindStr(typeValue valueKind) string {
	switch typeValue {
	case kindFloat64:
		return "gauge"
	case kindInt64:
		return "counter"
	default:
		return ""
	}
}

func (v *ValueMetric) DoUpdate(val ValueMetric) {
	switch v.kind {
	case kindFloat64:
		v.valueFloat = val.valueFloat
	case kindInt64:
		v.valueInt += val.valueInt
	default:
	}
}

func (v *ValueMetric) DoRead() ValueMetric {
	switch v.kind {
	case kindInt64:
		v.valueInt = 0
	default:
	}
	return *v
}

func (v *ValueMetric) KindOf(k valueKind) bool {
	return v.kind == k
}

func ConvertToFloatValueMetric(valF float64) *ValueMetric {
	val := new(ValueMetric)
	val.kind = kindFloat64
	val.valueFloat = valF
	return val
}

func ConvertToIntValueMetric(valI int64) *ValueMetric {
	val := new(ValueMetric)
	val.kind = kindInt64
	val.valueInt = valI
	return val
}

func ConvertToValueMetricInt(kind valueKind, delta *int64, value *float64) (*ValueMetric, error) {
	val := new(ValueMetric)
	val.kind = kind
	var err error
	switch kind {
	case kindFloat64:
		if value == nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}
		val.valueFloat = *value

	case kindInt64:
		if delta == nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}
		val.valueInt = *delta

	default:
		return nil, fmt.Errorf("failed %w : %w", ErrBadValue, ErrBadKindType)
	}
	return val, nil
}

func ConvertToValueMetric(kind valueKind, valstr string) (*ValueMetric, error) {
	val := new(ValueMetric)
	val.kind = kind
	var err error
	switch kind {
	case kindFloat64:
		val.valueFloat, err = strconv.ParseFloat(valstr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}

	case kindInt64:
		val.valueInt, err = strconv.ParseInt(valstr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed %w : %w", ErrBadValue, err)
		}

	default:
		return nil, fmt.Errorf("failed %w : %w", ErrBadValue, ErrBadKindType)
	}
	return val, nil
}

func ConvertValueMetricToPlain(val ValueMetric) (a, b string) {
	switch val.kind {
	case kindFloat64:
		a = GetKindStr(val.kind)
		b = strconv.FormatFloat(val.valueFloat, 'f', -1, 64)
	case kindInt64:
		a = GetKindStr(val.kind)
		b = strconv.FormatInt(val.valueInt, 10)
	default:
		a = "nan"
		b = "nan"
	}
	return a, b
}

func ConvertValueMetricToPlainOpt(val ValueMetric) (b string) {
	switch val.kind {
	case kindFloat64:

		b = strconv.FormatFloat(val.valueFloat, 'f', -1, 64)
	case kindInt64:

		b = strconv.FormatInt(val.valueInt, 10)
	default:
		b = "nan"
	}
	return b
}
