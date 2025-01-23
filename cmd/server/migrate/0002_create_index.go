package migrate

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { //nolint:gochecknoinits // for goose migration
	goose.AddMigrationContext(Up00002, Down00002)
}

func Up00002(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS metrics_test_idx ON metrics (name,kind);`)
	return err
}

func Down00002(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP INDEX metrics_test_idx;")
	return err
}
