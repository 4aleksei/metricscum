package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Address        string
	Level          string
	ReportInterval int64
	PollInterval   int64
	ContentJSON    bool
}

const AddressDefault string = ":8080"
const ReportIntervalDefault int64 = 10
const PollIntervalDefault int64 = 2
const LevelDefault string = "info"
const ContentJSONDefault bool = true

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "l", LevelDefault, "level logging")
	flag.Int64Var(&cfg.ReportInterval, "r", ReportIntervalDefault, "ReportInterval")
	flag.Int64Var(&cfg.PollInterval, "p", PollIntervalDefault, "PollInterval")
	flag.BoolVar(&cfg.ContentJSON, "j", ContentJSONDefault, "ContentJSON true/false")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			log.Fatalln("Error in converting env report interval to int: ", err)
		} else {
			cfg.ReportInterval = int64(val)
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			log.Fatalln("Error in converting env report interval to int: ", err)
		} else {
			cfg.PollInterval = int64(val)
		}
	}
	return cfg
}
