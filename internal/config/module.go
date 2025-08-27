package config

import (
	"fmt"
	"os"

	"go.uber.org/fx"
)

// Module provides *Config loaded from a YAML file.
var Module = fx.Provide(LoadFromEnv)

// LoadFromEnv loads configuration from LETLLM_CONFIG or defaults to "config.yaml".
func LoadFromEnv() (*Config, error) {
	path := os.Getenv("LETLLM_CONFIG")
	if path == "" {
		path = "config.yaml"
	}
	cfg, err := Load(path)
	if err != nil {
		return nil, fmt.Errorf("load config from %s: %w", path, err)
	}
	return cfg, nil
}
