// Package config - agent config read from  flags params and envVariables
package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

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
	ConfigJsonFile string
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
	ConfigDefaultJson     string = ""
	PublicKeyDefault      string = ""
)

func initDefaultCfg() *Config {
	cfg := new(Config)
	cfg.Address = AddressDefault
	cfg.Level = LevelDefault
	cfg.Key = KeyDefault
	cfg.ConfigJsonFile = ConfigDefaultJson

	cfg.ReportInterval = ReportIntervalDefault
	cfg.PollInterval = PollIntervalDefault
	cfg.ContentBatch = ContentBatchDefault

	cfg.ContentJSON = ContentJSONDefault

	cfg.PublicKeyFile = PublicKeyDefault
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

func NewConfig(l *logger.Logger) (*Config, error) {
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
	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")
	flag.StringVar(&cfg.Level, "v", cfg.Level, "level logging")
	flag.Int64Var(&cfg.ReportInterval, "r", cfg.ReportInterval, "ReportInterval")
	flag.Int64Var(&cfg.PollInterval, "p", cfg.PollInterval, "PollInterval")
	flag.BoolVar(&cfg.ContentJSON, "j", cfg.ContentJSON, "ContentJSON true/false")

	flag.Int64Var(&cfg.ContentBatch, "b", cfg.ContentBatch, "ContentBatch size uint")

	flag.StringVar(&cfg.Key, "k", cfg.Key, "key for signature")

	flag.Int64Var(&cfg.RateLimit, "l", cfg.RateLimit, "RateLimit, pool workers")

	flag.StringVar(&cfg.PublicKeyFile, "crypto-key", cfg.PublicKeyFile, "Public key file name")

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
			return nil, err
		} else {
			cfg.ReportInterval = int64(val)
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			l.L.Debug("Error in converting env report interval to int:", zap.Error(err))
			return nil, err
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
	return cfg, nil
}
