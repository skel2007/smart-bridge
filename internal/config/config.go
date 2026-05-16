package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultTuyaEndpoint = "https://openapi.tuyaeu.com"

type Config struct {
	Tuya TuyaConfig `yaml:"tuya"`
}

type TuyaConfig struct {
	Endpoint     string `yaml:"endpoint"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("config file not found: %s", path)
		}
		return Config{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid YAML in config file %s: %w", path, err)
	}

	applyDefaults(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if strings.TrimSpace(cfg.Tuya.Endpoint) == "" {
		cfg.Tuya.Endpoint = DefaultTuyaEndpoint
	}
}

func validate(cfg Config) error {
	if strings.TrimSpace(cfg.Tuya.ClientID) == "" {
		return errors.New("tuya.client_id is required")
	}
	if strings.TrimSpace(cfg.Tuya.ClientSecret) == "" {
		return errors.New("tuya.client_secret is required")
	}

	return nil
}
