package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewLogger(t *testing.T) {
	t.Run("Test NewLogger", func(t *testing.T) {
		var cfg Config
		cfg.Level = "debug"
		got := NewLogger(cfg)
		assert.NotNil(t, got)

		s := [7]string{"info", "error", "warn", "dpanic", "panic", "fatal", "debug"}

		for _, sval := range s {
			got.SetLevel(sval)
			assert.Equal(t, got.slevel, sval)
		}

		err := got.Start(context.Background())
		assert.Nil(t, err)
		errS := got.Stop(context.Background())
		assert.NotNil(t, errS)
	})
}

func Test_NewLog(t *testing.T) {
	t.Run("Test NewLog", func(t *testing.T) {
		got, err := NewLog("debug")
		assert.NotNil(t, got)
		assert.Nil(t, err)
	})
}
