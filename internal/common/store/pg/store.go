package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/4aleksei/metricscum/internal/common/store"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	DB struct {
		dbpool *pgxpool.Pool
	}

	Config struct {
		DatabaseDSN string
	}
)

func ProbePG(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}

func NewDB(cfg Config) (*DB, error) {
	if cfg.DatabaseDSN == "" {
		return &DB{dbpool: nil}, nil
	}

	var db *pgxpool.Pool
	ctx := context.Background()
	ctxB, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	err := utils.RetryAction(ctxB, utils.RetryTimes(), func(ctx context.Context) error {
		var err error
		db, err = pgxpool.New(ctx, cfg.DatabaseDSN)
		return err
	})

	if err != nil {
		return nil, err
	}

	ctxTimeOutPing, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	err = utils.RetryAction(ctxTimeOutPing, utils.RetryTimes(), func(ctx context.Context) error {
		ctxTime, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		return db.Ping(ctxTime)
	}, ProbePG)

	if err != nil {
		return nil, err
	}
	return &DB{dbpool: db}, nil
}

func (d *DB) Ping(ctx context.Context) error {
	return d.dbpool.Ping(ctx)
}

func (d *DB) Close(ctx context.Context) {
	d.dbpool.Close()
}

const (
	queryDefault             = `INSERT INTO metrics (name, kind, delta, value , updated_at) VALUES ($1,$2,$3,$4,now())`
	onConflictStatementDelta = ` ON CONFLICT (name, kind) 
		DO UPDATE SET delta=metrics.delta+excluded.delta,  updated_at = now() RETURNING name, kind, delta, value`
	onConflictStatementValue = ` ON CONFLICT (name, kind) 
		DO UPDATE SET  value=excluded.value , updated_at = now() RETURNING name, kind, delta, value`
)

func (d *DB) Upsert(ctx context.Context, modval store.Metrics, prog func(n string, k int, d int64, v float64) error) error {
	query := queryDefault
	if modval.Delta.Valid {
		query += onConflictStatementDelta
	} else {
		query += onConflictStatementValue
	}

	row := d.dbpool.QueryRow(ctx, query, modval.Name, modval.Kind, modval.Delta, modval.Value)
	if row != nil {
		var m store.Metrics
		err := row.Scan(&m.Name, &m.Kind, &m.Delta, &m.Value)
		if err != nil {
			return err
		}
		errP := prog(m.Name, m.Kind, m.Delta.Int64, m.Value.Float64)
		if errP != nil {
			return errP
		}
	} else {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) Upserts(ctx context.Context,
	modval []store.Metrics,
	limitbatch int, prog func(n string, k int, d int64, v float64) error) error {
	conn, err := d.dbpool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error begin tx: %w", err)
	}

	defer func() {
		defer func() { _ = tx.Rollback(ctx) }()
	}()

	var indexLimit int
	if len(modval) > limitbatch {
		indexLimit = limitbatch
	} else {
		indexLimit = len(modval)
	}

	for index := 0; index < len(modval); index += indexLimit {
		if (index + indexLimit) > len(modval) {
			indexLimit = len(modval) - index
		}

		batch := &pgx.Batch{}
		for i := 0; i < indexLimit; i++ {
			query := queryDefault
			if modval[i+index].Delta.Valid {
				query += onConflictStatementDelta
			} else {
				query += onConflictStatementValue
			}

			batch.Queue(query, modval[i+index].Name, modval[i+index].Kind, modval[i+index].Delta, modval[i+index].Value)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close() //nolint:gocritic // we are closing all batch results at end of loop

		for {
			row := br.QueryRow()
			var m store.Metrics
			err := row.Scan(&m.Name, &m.Kind, &m.Delta, &m.Value)
			if err != nil {
				break
			}
			errP := prog(m.Name, m.Kind, m.Delta.Int64, m.Value.Float64)
			if errP != nil {
				return errP
			}
		}
		if e := br.Close(); e != nil {
			return fmt.Errorf("closing batch result: %w", e)
		}
	}
	return tx.Commit(ctx)
}

func (d *DB) SelectValue(ctx context.Context, name string, prog func(n string, k int, d int64, v float64) error) error {
	row := d.dbpool.QueryRow(ctx, "SELECT name , kind , delta , value  FROM  metrics WHERE name=$1 LIMIT 1", name)
	if row != nil {
		var m store.Metrics
		err := row.Scan(&m.Name, &m.Kind, &m.Delta, &m.Value)
		if err != nil {
			return err
		}
		errP := prog(m.Name, m.Kind, m.Delta.Int64, m.Value.Float64)
		if errP != nil {
			return errP
		}
	} else {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) SelectValueAll(ctx context.Context, prog func(n string, k int, d int64, v float64) error) error {
	rows, err := d.dbpool.Query(ctx, "SELECT name , kind , delta , value  FROM metrics")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var m store.Metrics
		err := rows.Scan(&m.Name, &m.Kind, &m.Delta, &m.Value)
		if err != nil {
			return err
		}

		errK := prog(m.Name, m.Kind, m.Delta.Int64, m.Value.Float64)
		if errK != nil {
			return errK
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
