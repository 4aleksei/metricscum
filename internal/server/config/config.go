package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
	Level   string
}

const AddressDefault string = ":8080"
const LevelDefault string = "debug"

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "l", LevelDefault, "level of logging")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	return cfg
}
