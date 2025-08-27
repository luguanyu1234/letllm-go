package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration loaded from a YAML file.
type Config struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`

	// Route model names to a provider by prefix match (first match wins).
	// Example:
	// routes:
	//   - prefix: "gpt-"
	//     provider: "openai"
	//   - prefix: "gemini-"
	//     provider: "gemini"
	Routes []Route `yaml:"routes"`

	// Provider settings
	OpenAI struct {
		APIKey    string `yaml:"api_key"`
		DefaultModel string `yaml:"default_model"`
	} `yaml:"openai"`

	Gemini struct {
		APIKey    string `yaml:"api_key"`
		DefaultModel string `yaml:"default_model"`
	} `yaml:"gemini"`
}

type Route struct {
	Prefix   string `yaml:"prefix"`
	Provider string `yaml:"provider"` // "openai" or "gemini"
}

// Load loads configuration from the provided file path.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	// Allow env override for API keys
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.OpenAI.APIKey = v
	}
	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		cfg.Gemini.APIKey = v
	}
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}
	return &cfg, nil
}
