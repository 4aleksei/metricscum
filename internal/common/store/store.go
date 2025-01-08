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
		DATABASE_DSN string
	}
)

const (
	//DATABASE_DSN_DEFAULT string = "host=localhost user=metrics dbname=dbname password=metricspassword  sslmode=disable"

	DATABASE_DSN_DEFAULT string = "postgresql://localhost/dbname?user=metrics&password=metricspassword"
)

func NewDB(cfg Config) (*DB, error) {
	//db, err := sqlx.Connect("postgres", "user=metrics dbname=dbname  sslmode=disable")
	//"host=localhost user=metrics dbname=dbname password=metricspassword  sslmode=disable"
	db, err := sqlx.Open("pgx", cfg.DATABASE_DSN)
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
	flag.StringVar(&cfg.DATABASE_DSN, "d", DATABASE_DSN_DEFAULT, "DATABASE_DSN")
}

func ReadConfigEnv(cfg *Config) {
	if envDBADDR := os.Getenv("DATABASE_DSN"); envDBADDR != "" {
		cfg.DATABASE_DSN = envDBADDR
	}
}
