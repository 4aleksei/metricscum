package valuemetric

import (
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
