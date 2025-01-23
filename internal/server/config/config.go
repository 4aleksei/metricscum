package config

import (
	"flag"
	"os"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/store/pg"
)

type Config struct {
	Address  string
	Level    string
	FilePath string

	Repcfg repository.Config
	DBcfg  pg.Config
}

const (
	AddressDefault     string = ":8080"
	LevelDefault       string = "debug"
	FilePathDefault    string = "./data.store"
	databaseDSNDefault string = ""
)

func readConfigFlag(cfg *pg.Config) {
	flag.StringVar(&cfg.DatabaseDSN, "d", databaseDSNDefault, "DATABASE_DSN")
}

func readConfigEnv(cfg *pg.Config) {
	if envDBADDR := os.Getenv("DATABASE_DSN"); envDBADDR != "" {
		cfg.DatabaseDSN = envDBADDR
	}
}

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "l", LevelDefault, "level of logging")
	flag.StringVar(&cfg.FilePath, "f", FilePathDefault, "FilePath store")

	repository.ReadConfigFlag(&cfg.Repcfg)
	readConfigFlag(&cfg.DBcfg)

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FilePath = envFilePath
	}
	repository.ReadConfigEnv(&cfg.Repcfg)
	readConfigEnv(&cfg.DBcfg)

	return cfg
}
