package store

import (
	"context"
	"flag"
	"os"

	//_ "github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"

	//"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	DB struct {
		DB *sqlx.DB
	}

	Config struct {
		databaseDSN string
	}
)

const (
	//DATABASE_DSN_DEFAULT string = "host=localhost user=metrics dbname=dbname password=metricspassword  sslmode=disable"
	//postgresql://localhost/dbname?user=metrics&password=metricspassword
	databaseDSNDefault string = ""
)

func NewDB(cfg Config) (*DB, error) {
	if cfg.databaseDSN == "" {
		return &DB{DB: nil}, nil
	}
	db, err := sqlx.Open("pgx", cfg.databaseDSN)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

func ReadConfigFlag(cfg *Config) {
	flag.StringVar(&cfg.databaseDSN, "d", databaseDSNDefault, "DATABASE_DSN")
}

func ReadConfigEnv(cfg *Config) {
	if envDBADDR := os.Getenv("DATABASE_DSN"); envDBADDR != "" {
		cfg.databaseDSN = envDBADDR
	}
}
