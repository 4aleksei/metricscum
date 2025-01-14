package config

import (
	"flag"
	"os"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/store"
)

type Config struct {
	Address  string
	Level    string
	FilePath string

	Repcfg repository.Config
	DBcfg  store.Config
}

const (
	AddressDefault  string = ":8080"
	LevelDefault    string = "debug"
	FilePathDefault string = "./data.store"
)

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "l", LevelDefault, "level of logging")
	flag.StringVar(&cfg.FilePath, "f", FilePathDefault, "FilePath store")

	repository.ReadConfigFlag(&cfg.Repcfg)
	store.ReadConfigFlag(&cfg.DBcfg)

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FilePath = envFilePath
	}
	repository.ReadConfigEnv(&cfg.Repcfg)
	store.ReadConfigEnv(&cfg.DBcfg)

	return cfg
}
