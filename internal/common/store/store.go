package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrConflict = errors.New("data conflict")

type Store interface {
	Upsert(ctx context.Context, val Metrics, prog func(n string, k int, d int64, v float64) error) error
	Upserts(ctx context.Context, vals []Metrics, lim int, prog func(n string, k int, d int64, v float64) error) error
	SelectValue(ctx context.Context, name string, prog func(n string, k int, d int64, v float64) error) error
	SelectValueAll(ctx context.Context, prog func(n string, k int, d int64, v float64) error) error
	Close(ctx context.Context)
	Ping(ctx context.Context) error
}

type (
	Metrics struct {
		Name  string          `db:"name"`
		Kind  int             `db:"kind"`
		Delta sql.NullInt64   `db:"delta"`
		Value sql.NullFloat64 `db:"value"`
	}
)
