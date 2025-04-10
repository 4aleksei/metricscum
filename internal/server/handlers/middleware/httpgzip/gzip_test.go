package httpgzip

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/4aleksei/metricscum/internal/common/streams/compressors/zipdata"
	"github.com/stretchr/testify/assert"
)

func Test_NewReader(t *testing.T) {
	writer := zipdata.NewWriter()
	var buf bytes.Buffer
	w, _ := writer.OpenWriter(io.Writer(&buf))
	_, _ = w.Write([]byte("dfdfdfdfasasasasasasasaffff"))
	_ = writer.CloseWrite()

	t.Run("Test NewReader", func(t *testing.T) {
		got, err := NewCompressReader(io.NopCloser(strings.NewReader(buf.String())))
		assert.Nil(t, err)
		assert.NotNil(t, got)
		p := make([]byte, 100)
		n, errR := got.Read(p)
		assert.Nil(t, errR)
		assert.NotNil(t, n)

		errC := got.Close()
		assert.Nil(t, errC)
	})
}
func Test_NewWriter(t *testing.T) {
	t.Run("Test NewWriter", func(t *testing.T) {
		var w http.ResponseWriter
		got := NewCompressWriter(w)
		assert.NotNil(t, got)
		assert.NotNil(t, got.zw)
	})
}
