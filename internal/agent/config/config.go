package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Address        string
	ReportInterval uint
	PollInterval   uint
}

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")

	flag.UintVar(&cfg.ReportInterval, "r", 10, "ReportInterval")
	flag.UintVar(&cfg.PollInterval, "p", 2, "PollInterval")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {

		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			fmt.Println("Error in converting env report interval to int: ", err)
		} else {
			cfg.ReportInterval = uint(val)
		}

	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {

		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			fmt.Println("Error in converting env report interval to int: ", err)
		} else {
			cfg.PollInterval = uint(val)
		}
	}

	return cfg
}
