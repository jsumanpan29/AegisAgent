package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Modules           []string `yaml:"modules"`
	LogPath           string   `yaml:"log_path"`
	HeartbeatInterval int      `yaml:"heartbeat_interval"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(file, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
