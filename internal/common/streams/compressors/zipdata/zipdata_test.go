package zipdata

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewReader(t *testing.T) {
	t.Run("Test NewReader", func(t *testing.T) {
		assert.NotNil(t, NewReader())
	})
}
func Test_NewWriter(t *testing.T) {
	t.Run("Test NewWriter", func(t *testing.T) {
		assert.NotNil(t, NewWriter())
	})
}

func Test_OpenWrite(t *testing.T) {
	t.Run("Test Test_OpenWrite", func(t *testing.T) {
		writer := NewWriter()
		var buf bytes.Buffer
		w, err := writer.OpenWriter(io.Writer(&buf))
		assert.Nil(t, err)
		n, errC := w.Write([]byte("dfdfdfdfasasasasasasasaffff"))
		assert.Nil(t, errC)
		assert.NotNil(t, n)
		errCl := writer.CloseWrite()
		assert.Nil(t, errCl)
	})
}

func Test_OpenReader(t *testing.T) {
	t.Run("Test OpenReader Error", func(t *testing.T) {
		reader := NewReader()
		r, err := reader.OpenReader(strings.NewReader("frewd"))
		assert.NotNil(t, err)
		assert.Nil(t, r)
	})

	writer := NewWriter()
	var buf bytes.Buffer
	w, _ := writer.OpenWriter(io.Writer(&buf))
	_, _ = w.Write([]byte("dfdfdfdfasasasasasasasaffff"))
	_ = writer.CloseWrite()

	t.Run("Test OpenReader", func(t *testing.T) {
		reader := NewReader()
		r, err := reader.OpenReader(strings.NewReader(buf.String()))
		assert.Nil(t, err)
		_, errC := io.ReadAll(r)
		assert.Nil(t, errC)
		errCl := reader.CloseRead()
		assert.Nil(t, errCl)
	})
}
