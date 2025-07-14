// Package config json decode file
package config

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

type Duration time.Duration

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
	Address        *string   `json:"address,omitempty"`
	ReportInterval *Duration `json:"report_interval,omitempty"`
	PollInterval   *Duration `json:"poll_interval,omitempty"`
	PublicKeyFile  *string   `json:"crypto_key,omitempty"`
	Level          *string   `json:"level,omitempty"`
	Key            *string   `json:"key,omitempty"`
	ContentBatch   *int64    `json:"content_batch,omitempty"`
	RateLimit      *int64    `json:"rate_limit,omitempty"`
	ContentJSON    *bool     `json:"content_json,omitempty"`
	Grpc           *bool     `json:"grpc,omitempty"`
	CertFile       *string   `json:"crypto_cert,omitempty"`
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
	if jsonconfig.ContentJSON != nil {
		cfg.ContentJSON = *jsonconfig.ContentJSON
	}

	if jsonconfig.PollInterval != nil {
		cfg.PollInterval = int64(*jsonconfig.PollInterval) / 1000000000
	}

	if jsonconfig.ReportInterval != nil {
		cfg.ReportInterval = int64(*jsonconfig.ReportInterval) / 1000000000
	}

	if jsonconfig.ContentBatch != nil {
		cfg.ContentBatch = *jsonconfig.ContentBatch
	}

	if jsonconfig.Address != nil {
		cfg.Address = *jsonconfig.Address
	}

	if jsonconfig.Key != nil {
		cfg.Key = *jsonconfig.Key
	}

	if jsonconfig.PublicKeyFile != nil {
		cfg.PublicKeyFile = *jsonconfig.PublicKeyFile
	}

	if jsonconfig.Level != nil {
		cfg.Level = *jsonconfig.Level
	}
	if jsonconfig.RateLimit != nil {
		cfg.RateLimit = *jsonconfig.RateLimit
	}

	if jsonconfig.Grpc != nil {
		cfg.Grpc = *jsonconfig.Grpc
	}

	if jsonconfig.CertFile != nil {
		cfg.CertKeyFile = *jsonconfig.CertFile
	}

	return nil
}
