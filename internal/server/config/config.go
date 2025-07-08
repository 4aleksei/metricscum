// Package config
package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/common/store/pg"
)

type Config struct {
	Address        string
	Level          string
	FilePath       string
	DBcfg          pg.Config
	Key            string
	Repcfg         repository.Config
	PrivateKeyFile string
	ConfigJsonFile string
	Cidr           string
}

const (
	AddressDefault        string = ":8080"
	LevelDefault          string = "debug"
	FilePathDefault       string = "./data.store"
	databaseDSNDefault    string = ""
	KeyDefault            string = ""
	ConfigDefaultJson     string = ""
	WriteIntervalDefault  int64  = 300
	RestoreDefault        bool   = true
	PrivateKeyFileDefault string = ""
	CidrDefault                  = ""
)

func initDefaultCfg() *Config {
	cfg := new(Config)
	cfg.Address = AddressDefault
	cfg.Level = LevelDefault
	cfg.FilePath = FilePathDefault
	cfg.DBcfg.DatabaseDSN = databaseDSNDefault
	cfg.Key = KeyDefault
	cfg.Repcfg.Restore = RestoreDefault
	cfg.Repcfg.Interval = WriteIntervalDefault
	cfg.ConfigJsonFile = ConfigDefaultJson
	cfg.PrivateKeyFile = PrivateKeyFileDefault
	cfg.Cidr = CidrDefault
	return cfg
}

func getJsonFileName(key string, cfg *Config) bool {
	for i, value := range os.Args {
		if value == key {
			if (i + 1) < len(os.Args) {
				cfg.ConfigJsonFile = strings.Trim(os.Args[i+1], " '\"")
			}
		}
		if value == "-h" || value == "--help" {
			cfg.ConfigJsonFile = ""
			return false
		}
	}
	return true
}

func readConfigFlagPg(cfg *pg.Config) {
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DATABASE_DSN")
}

func readConfigEnvPg(cfg *pg.Config) {
	if envDBADDR := os.Getenv("DATABASE_DSN"); envDBADDR != "" {
		cfg.DatabaseDSN = envDBADDR
	}
}

func readConfigFlagRep(cfg *repository.Config) {
	flag.Int64Var(&cfg.Interval, "i", cfg.Interval, "Write data Interval")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore data true/false")
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

func NewConfig() (*Config, error) {
	cfg := initDefaultCfg()

	dnhelp := getJsonFileName("-c", cfg)
	if envConfig := os.Getenv("CONFIG"); envConfig != "" && dnhelp {
		cfg.ConfigJsonFile = envConfig
	}

	if cfg.ConfigJsonFile != "" {
		err := loadConfigJson(cfg.ConfigJsonFile, cfg)
		if err != nil {
			return nil, err
		}
	}
	flag.StringVar(&cfg.ConfigJsonFile, "c", cfg.ConfigJsonFile, "Config file name in json format")

	flag.StringVar(&cfg.Cidr, "t", cfg.Cidr, "Trusted subnet (CIDR)")

	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")

	flag.StringVar(&cfg.Level, "v", cfg.Level, "level of logging")
	flag.StringVar(&cfg.FilePath, "f", cfg.FilePath, "FilePath store")

	readConfigFlagRep(&cfg.Repcfg)
	readConfigFlagPg(&cfg.DBcfg)

	flag.StringVar(&cfg.Key, "k", cfg.Key, "key for signature")
	flag.StringVar(&cfg.PrivateKeyFile, "crypto-key", cfg.PrivateKeyFile, "Private key file name (pem)")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
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

	if envTrustNet := os.Getenv("TRUSTED_SUBNET"); envTrustNet != "" {
		cfg.Cidr = envTrustNet
	}

	readConfigEnvRep(&cfg.Repcfg)
	readConfigEnvPg(&cfg.DBcfg)

	return cfg, nil
}
