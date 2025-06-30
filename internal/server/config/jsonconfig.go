// Package config
package config

import (
	"encoding/json"
	"io"

	//	"strconv"
	"fmt"
	"os"
	"time"
)

type Jsonconfig struct {
	Address       *string        `json:"address,omitempty"`
	Restore       *bool          `json:"restore,omitempty"`
	StoreInterval *time.Duration `json:"store_interval,omitempty"`
	StoreFile     *string        `json:"store_file,omitempty"`
	DatabaseDsn   *string        `json:"database_dsn,omitempty"`
	CryptoKey     *string        `json:"crypto_key,omitempty"`
}

func jsonConfigDecode(body io.ReadCloser) (*Jsonconfig, error) {
	dec := json.NewDecoder(body)
	var config Jsonconfig
	err := dec.Decode(&config)
	return &config, err
}

func loadConfigJson(name string, cfg *Config) error {
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
	}()

	jsonconfig, err := jsonConfigDecode(file)
	if err != nil {
		return err
	}

	if jsonconfig.Address != nil {
		cfg.Address = *jsonconfig.Address
	}
	if jsonconfig.Restore != nil {
		cfg.Repcfg.Restore = *jsonconfig.Restore
	}

	return nil
}
