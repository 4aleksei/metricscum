// This is custom goose binary with sqlite3 support only.

package migrate

import (
	"context"

	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type gooseLogger struct {
	l *zap.Logger
}

func (l *gooseLogger) Fatalf(format string, v ...interface{}) {
	l.l.Fatal("goose fatal", zap.String("msg", fmt.Sprintf(format, v...)))
}
func (l *gooseLogger) Printf(format string, v ...interface{}) {
	l.l.Info("goose info", zap.String("msg", fmt.Sprintf(format, v...)))
}

func Migrate(l *zap.Logger, dbstring string, command string) error {
	var g = gooseLogger{l: l}

	goose.SetLogger(&g)

	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {

		}

	}()

	arguments := []string{}
	dir := new(string)
	*dir = "."
	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, *dir, arguments...); err != nil {
		return err
	}
	return nil
}
