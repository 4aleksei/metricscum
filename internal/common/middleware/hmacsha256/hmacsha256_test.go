package hmacsha256

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewWriter(t *testing.T) {
	t.Run("Test NewWriter Write Get Sig", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(io.Writer(&buf), []byte("keykey"))
		assert.NotNil(t, w.w)
		n, err := w.Write([]byte("DataData"))
		assert.Nil(t, err)

		assert.NotNil(t, n)

		b := w.GetSig()

		assert.NotNil(t, b)
	})
}

func Test_NewReader(t *testing.T) {
	t.Run("Test NewReader Read Get Sig", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader("sdksdskdskjdsdsd"))
		r := NewReader(body, []byte("keykey"))
		assert.NotNil(t, r.hr)
		rB := make([]byte, 100)
		n, err := r.Read(rB)
		assert.Nil(t, err)

		assert.NotNil(t, n)

		b := r.GetSig()
		assert.NotNil(t, b)
		bG, errC := GetSig(r)
		assert.NotNil(t, bG)
		assert.Nil(t, errC)
		errCC := r.Close()

		assert.Nil(t, errCC)
	})
}
