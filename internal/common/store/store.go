package store

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type (
	DB struct {
		DB *sqlx.DB
	}

	Config struct {
		DatabaseDSN string
	}

	Metrics struct {
		Name  string          `db:"name"`
		Kind  int             `db:"kind"`
		Delta sql.NullInt64   `db:"delta"`
		Value sql.NullFloat64 `db:"value"`
	}

	TX struct {
		tx *sqlx.Tx
	}
)

const (
	databaseDSNDefault string = ""
)

func NewDB(cfg Config) (*DB, error) {
	if cfg.DatabaseDSN == "" {
		return &DB{DB: nil}, nil
	}
	db, err := sqlx.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	ctxTimeOutPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = db.PingContext(ctxTimeOutPing)
	if err != nil {
		return nil, err
	}
	ctxTimeOut, cancel2 := context.WithTimeout(ctx, 3*time.Second)
	defer cancel2()
	err = createMetricsTable(ctxTimeOut, db)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

func createMetricsTable(ctx context.Context, db *sqlx.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, errE := tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS metrics (
            name varchar(128)  not null,
            kind int4 not null ,
			delta bigint,
            value double precision,
            updated_at timestamptz not null DEFAULT NOW(),
			primary key(name, kind)
        )
    `)
	if errE != nil {
		return errE
	}
	_, errE = tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS metrics_idx ON metrics (name,kind)`)
	if errE != nil {
		return errE
	}
	return tx.Commit()
}

func ReadConfigFlag(cfg *Config) {
	flag.StringVar(&cfg.DatabaseDSN, "d", databaseDSNDefault, "DATABASE_DSN")
}

func ReadConfigEnv(cfg *Config) {
	if envDBADDR := os.Getenv("DATABASE_DSN"); envDBADDR != "" {
		cfg.DatabaseDSN = envDBADDR
	}
}

func (d *DB) BeginTx() (*TX, error) {
	tx, err := d.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return &TX{tx: tx}, nil
}

func setint64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func setfloat64(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

func (tx *TX) Upsert(name string, kind int, delta *int64, value *float64, prog func(n string, k int, d int64, v float64) error) error {
	var m = Metrics{
		Name:  name,
		Kind:  kind,
		Delta: sql.NullInt64{Valid: delta != nil, Int64: setint64(delta)},
		Value: sql.NullFloat64{Valid: value != nil, Float64: setfloat64(value)},
	}

	var onConflictStatement string
	sqlStr := `INSERT INTO metrics (name, kind, delta, value , updated_at) VALUES (:name,:kind,:delta,:value,now())`
	if delta != nil {
		onConflictStatement = ` ON CONFLICT (name, kind) 
		DO UPDATE SET delta=metrics.delta+excluded.delta,  updated_at = now() RETURNING name, kind, delta, value`
	} else {
		onConflictStatement = ` ON CONFLICT (name, kind) 
		DO UPDATE SET  value=excluded.value , updated_at = now() RETURNING name, kind, delta, value`
	}
	query, queryArgs, errB := tx.tx.BindNamed(sqlStr, m)
	if errB != nil {
		_ = tx.tx.Rollback()
		return errB
	}
	query = tx.tx.Rebind(query)
	query += onConflictStatement
	rows, err := tx.tx.Queryx(query, queryArgs...)
	if err != nil {
		_ = tx.tx.Rollback()
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.StructScan(&m)
		if err != nil {
			return err
		}
		errP := prog(m.Name, m.Kind, m.Delta.Int64, m.Value.Float64)
		if errP != nil {
			return errP
		}
	} else {
		_ = tx.tx.Rollback()
		return sql.ErrNoRows
	}
	return nil
}

func (tx *TX) EndTx() error {
	err := tx.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) SelectValue(name string, prog func(n string, k int, d int64, v float64) error) error {
	row := d.DB.QueryRowx("SELECT name , kind , delta , value  FROM  metrics WHERE name=$1 LIMIT 1", name)
	if row != nil {
		var m Metrics
		err := row.StructScan(&m)
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

func (d *DB) SelectValueAll(prog func(n string, k int, d int64, v float64) error) error {
	rows, err := d.DB.Queryx("SELECT name , kind , delta , value  FROM metrics")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var m Metrics
		err := rows.StructScan(&m)
		if err != nil {
			return err
		}

		errK := prog(m.Name, m.Kind, m.Delta.Int64, m.Value.Float64)
		if errK != nil {
			return errK
		}
	}
	return nil
}
