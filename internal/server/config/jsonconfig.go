// Package config json decode file
package config

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

type (
	Duration time.Duration
)

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, err := time.ParseDuration(str)
	*d = Duration(val)
	return err
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

type Jsonconfig struct {
	Restore       *bool     `json:"restore,omitempty"`
	StoreInterval *Duration `json:"store_interval,omitempty"`

	Ncidr *string `json:"trusted_subnet,omitempty"`

	DatabaseDsn *string `json:"database_dsn,omitempty"`

	Address   *string `json:"address,omitempty"`
	StoreFile *string `json:"store_file,omitempty"`
	CryptoKey *string `json:"crypto_key,omitempty"`
	Key       *string `json:"key,omitempty"`
	Level     *string `json:"level,omitempty"`
}

func jsonConfigDecode(body io.ReadCloser) (*Jsonconfig, error) {
	dec := json.NewDecoder(body)
	var config Jsonconfig
	err := dec.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func loadConfigJson(name string, cfg *Config) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	jsonconfig, err := jsonConfigDecode(file)
	if err != nil {
		return err
	}
	if jsonconfig.Restore != nil {
		cfg.Repcfg.Restore = *jsonconfig.Restore
	}

	if jsonconfig.StoreInterval != nil {
		cfg.Repcfg.Interval = int64(*jsonconfig.StoreInterval) / 1000000000
	}
	if jsonconfig.DatabaseDsn != nil {
		cfg.DBcfg.DatabaseDSN = *jsonconfig.DatabaseDsn
	}

	if jsonconfig.Address != nil {
		cfg.Address = *jsonconfig.Address
	}

	if jsonconfig.Ncidr != nil {
		cfg.Cidr = *jsonconfig.Ncidr
	}

	if jsonconfig.StoreFile != nil {
		cfg.FilePath = *jsonconfig.StoreFile
	}

	if jsonconfig.Key != nil {
		cfg.Key = *jsonconfig.Key
	}

	if jsonconfig.CryptoKey != nil {
		cfg.PrivateKeyFile = *jsonconfig.CryptoKey
	}

	if jsonconfig.Level != nil {
		cfg.Level = *jsonconfig.Level
	}
	return nil
}
