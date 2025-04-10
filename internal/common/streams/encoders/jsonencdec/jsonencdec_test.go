package jsonencdec

import (
	"bytes"

	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"io"

	"github.com/4aleksei/metricscum/internal/common/models"
)

func Test_NewReader(t *testing.T) {
	t.Run("Test Json NewReader", func(t *testing.T) {
		assert.NotNil(t, NewReader())
	})
}
func Test_NewWriter(t *testing.T) {
	t.Run("Test Json NewWriter", func(t *testing.T) {
		assert.NotNil(t, NewWriter())
	})
}

func Test_OpenWrite(t *testing.T) {
	var vI int64 = 10
	a := models.Metrics{
		ID:    "Test",
		MType: "counter",
		Delta: &vI,
	}

	t.Run("Test Json Test_OpenWrite", func(t *testing.T) {
		writer := NewWriter()
		var buf bytes.Buffer
		writer.OpenWriter(io.Writer(&buf))
		assert.NotNil(t, writer.encoder)
		errC := writer.WriteData(&a)
		assert.Nil(t, errC)
		writer.CloseWrite()
		assert.Nil(t, writer.encoder)
	})
}

func Test_OpenReader(t *testing.T) {
	var m models.Metrics
	t.Run("Test Json Test_OpenRead", func(t *testing.T) {
		reader := NewReader()
		s := "{ \"id\":\"TestMetr\" , \"type\":\"counter\" , \"delta\":44  }"
		reader.OpenReader(strings.NewReader(s))
		assert.NotNil(t, reader.decoder)
		errC := reader.ReadData(&m)
		assert.Nil(t, errC)
		reader.CloseRead()
		assert.Nil(t, reader.decoder)
	})
}
