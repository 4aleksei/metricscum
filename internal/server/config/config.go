package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
}

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	return cfg
}
