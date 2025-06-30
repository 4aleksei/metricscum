// Package config - agent config read from  flags params and envVariables
package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"go.uber.org/zap"
)

type Config struct {
	Lcfg           *logger.Config
	Address        string
	Level          string
	Key            string
	PublicKeyFile  string
	ReportInterval int64
	PollInterval   int64
	ContentBatch   int64
	RateLimit      int64
	ContentJSON    bool
}

const (
	AddressDefault        string = ":8080"
	ReportIntervalDefault int64  = 10
	PollIntervalDefault   int64  = 2
	LevelDefault          string = "info"
	ContentJSONDefault    bool   = true
	ContentBatchDefault   int64  = 0
	KeyDefault            string = ""
	RateLimitDefault      int64  = 10
)

func GetConfig(l *logger.Logger) *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Address, "a", AddressDefault, "address and port to run server")
	flag.StringVar(&cfg.Level, "v", LevelDefault, "level logging")
	flag.Int64Var(&cfg.ReportInterval, "r", ReportIntervalDefault, "ReportInterval")
	flag.Int64Var(&cfg.PollInterval, "p", PollIntervalDefault, "PollInterval")
	flag.BoolVar(&cfg.ContentJSON, "j", ContentJSONDefault, "ContentJSON true/false")

	flag.Int64Var(&cfg.ContentBatch, "b", ContentBatchDefault, "ContentBatch size uint")

	flag.StringVar(&cfg.Key, "k", KeyDefault, "key for signature")

	flag.Int64Var(&cfg.RateLimit, "l", RateLimitDefault, "RateLimit, pool workers")

	flag.StringVar(&cfg.PublicKeyFile, "crypto-key", KeyDefault, "Public key file name")

	flag.Parse()

	cfg.Lcfg = new(logger.Config)
	cfg.Lcfg.Level = cfg.Level

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	if envPublicKey := os.Getenv("CRYPTO_KEY"); envPublicKey != "" {
		cfg.PublicKeyFile = envPublicKey
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			l.L.Debug("Error in converting env report interval to int:", zap.Error(err))
		} else {
			cfg.ReportInterval = int64(val)
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			l.L.Debug("Error in converting env report interval to int:", zap.Error(err))
		} else {
			cfg.PollInterval = int64(val)
		}
	}

	if !cfg.ContentJSON {
		cfg.ContentBatch = 0
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 1
	}

	return cfg
}
