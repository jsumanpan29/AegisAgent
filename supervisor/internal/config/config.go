package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Modules []string `json:"modules"`
	LogPath string   `json:"log_path"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
