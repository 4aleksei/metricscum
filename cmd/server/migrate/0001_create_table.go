package migrate

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00001, Down00001)
}

func Up00001(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS metrics (
            name varchar(128)  not null,
            kind int4 not null ,
			delta bigint,
            value double precision,
            updated_at timestamptz not null DEFAULT NOW(),
			primary key(name, kind)
        );
    `)
	return err
}

func Down00001(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE metrics_test;")
	return err
}
