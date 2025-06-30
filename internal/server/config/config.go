package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/store/pg"
)

const (
	WriteIntervalDefault int64 = 300
	RestoreDefault       bool  = true
)

type Config struct {
	Address        string
	Level          string
	FilePath       string
	DBcfg          pg.Config
	Key            string
	Repcfg         repository.Config
	PrivateKeyFile string
}

const (
	AddressDefault     string = ":8080"
	LevelDefault       string = "debug"
	FilePathDefault    string = "./data.store"
	databaseDSNDefault string = ""
	KeyDefault         string = ""
	ConfigFile         string = ""
)

func initDefaultCfg() *Config {
	cfg := new(Config)
	cfg.Address = AddressDefault
	cfg.Level = LevelDefault
	cfg.FilePath = FilePathDefault

	return cfg
}

func getJsonFileName() string {
	for i, value := range os.Args {
		if value == "-c" {
			if (i + 1) < len(os.Args) {
				return os.Args[i+1]
			}
		}
	}
	return ""
}

func readConfigFlagPg(cfg *pg.Config) {
	flag.StringVar(&cfg.DatabaseDSN, "d", databaseDSNDefault, "DATABASE_DSN")
}

func readConfigEnvPg(cfg *pg.Config) {
	if envDBADDR := os.Getenv("DATABASE_DSN"); envDBADDR != "" {
		cfg.DatabaseDSN = envDBADDR
	}
}

func readConfigFlagRep(cfg *repository.Config) {
	flag.Int64Var(&cfg.Interval, "i", WriteIntervalDefault, "Write data Interval")
	flag.BoolVar(&cfg.Restore, "r", RestoreDefault, "Restore data true/false")
}

func readConfigEnvRep(cfg *repository.Config) {
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		val, err := strconv.Atoi(envStoreInterval)
		if err == nil {
			if val >= 0 {
				cfg.Interval = int64(val)
			}
		}
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		switch envRestore {
		case "true":
			cfg.Restore = true

		case "false":
			cfg.Restore = false
		}
	}
}

func GetConfig() *Config {
	jsonName := getJsonFileName()
	cfg := initDefaultCfg()

	if jsonName != "" {
		loadConfigJson(jsonName, cfg)
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")

	flag.StringVar(&cfg.Level, "v", LevelDefault, "level of logging")
	flag.StringVar(&cfg.FilePath, "f", FilePathDefault, "FilePath store")

	readConfigFlagRep(&cfg.Repcfg)
	readConfigFlagPg(&cfg.DBcfg)

	flag.StringVar(&cfg.Key, "k", KeyDefault, "key for signature")
	flag.StringVar(&cfg.PrivateKeyFile, "crypto-key", KeyDefault, "Private key file")

	flag.Parse()
	fmt.Println(cfg.Address)

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr

		fmt.Println("SET ENV ADDRESS ", cfg.Address)

	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FilePath = envFilePath
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	if envPrivateKeyFile := os.Getenv("CRYPTO_KEY"); envPrivateKeyFile != "" {
		cfg.PrivateKeyFile = envPrivateKeyFile
	}

	readConfigEnvRep(&cfg.Repcfg)
	readConfigEnvPg(&cfg.DBcfg)

	return cfg
}
