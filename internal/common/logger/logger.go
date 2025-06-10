// Package logger
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Logger struct {
		L      *zap.Logger
		level  zap.AtomicLevel
		slevel string
	}

	Config struct {
		Level string
	}
)

func NewLogger(cfg Config) *Logger {
	z, lvl, _ := newLog(cfg.Level)
	return &Logger{L: z, slevel: cfg.Level, level: lvl}
}

func (l *Logger) SetLevel(lev string) {
	l.slevel = lev

	switch l.slevel {
	case "debug":
		l.level.SetLevel(zapcore.DebugLevel)
	case "info":
		l.level.SetLevel(zapcore.InfoLevel)
	case "error":
		l.level.SetLevel(zapcore.ErrorLevel)
	case "warn":
		l.level.SetLevel(zapcore.WarnLevel)
	case "dpanic":
		l.level.SetLevel(zapcore.DPanicLevel)
	case "panic":
		l.level.SetLevel(zapcore.PanicLevel)
	case "fatal":
		l.level.SetLevel(zapcore.FatalLevel)
	default:
	}
}

func (l *Logger) Start(ctx context.Context) error {
	l.SetLevel(l.slevel)
	return nil
}

func (l *Logger) Stop(ctx context.Context) error {
	return l.L.Sync()
}

func NewLog(level string) (*zap.Logger, error) {
	log, _, err := newLog(level)
	return log, err
}

func newLog(level string) (*zap.Logger, zap.AtomicLevel, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, lvl, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, lvl, err
	}
	return zl, lvl, nil
}
