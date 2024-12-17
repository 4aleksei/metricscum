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
	ReportInterval uint
	PollInterval   uint
	ContentJson    uint
}

const AddressDefault string = ":8080"
const ReportIntervalDefault uint = 10
const PollIntervalDefault uint = 2
const LevelDefault string = "info"
const ContentJsonDefault uint = 1

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "l", LevelDefault, "level logging")
	flag.UintVar(&cfg.ReportInterval, "r", ReportIntervalDefault, "ReportInterval")
	flag.UintVar(&cfg.PollInterval, "p", PollIntervalDefault, "PollInterval")
	flag.UintVar(&cfg.ContentJson, "j", ContentJsonDefault, "ContentJson 1-true 0-false")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {

		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			log.Fatalln("Error in converting env report interval to int: ", err)
		} else {
			cfg.ReportInterval = uint(val)
		}

	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {

		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			log.Fatalln("Error in converting env report interval to int: ", err)
		} else {
			cfg.PollInterval = uint(val)
		}
	}

	return cfg
}
