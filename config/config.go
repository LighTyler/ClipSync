package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Peer   []string `json:"addr"`
	Listen string   `json:"listen"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config - NewConfig - ReadFile: %w", err)
	}
	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config - NewConfig - Unmarshal: %w", err)
	}
	return cfg, nil
}
