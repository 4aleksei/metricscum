package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Address       string
	Level         string
	WriteInterval uint
	Restore       bool
	FilePath      string
}

const AddressDefault string = ":8080"
const LevelDefault string = "debug"
const WriteIntervalDefault uint = 300
const RestoreDefault bool = true
const FilePathDefault string = "./data.store"

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "l", LevelDefault, "level of logging")
	flag.UintVar(&cfg.WriteInterval, "i", WriteIntervalDefault, "Write data Interval")
	flag.BoolVar(&cfg.Restore, "r", RestoreDefault, "Restore data true/false")
	flag.StringVar(&cfg.FilePath, "f", FilePathDefault, "FilePath store")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		val, err := strconv.Atoi(envStoreInterval)
		if err == nil {
			if val >= 0 {
				cfg.WriteInterval = uint(val)
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

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FilePath = envFilePath
	}

	return cfg
}
