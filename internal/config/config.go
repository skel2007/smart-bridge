package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultTuyaEndpoint     = "https://openapi.tuyaeu.com"
	DefaultYandexPathPrefix = "/api/yandex"
)

type Config struct {
	HTTP   HTTPConfig   `yaml:"http"`
	Tuya   TuyaConfig   `yaml:"tuya"`
	Yandex YandexConfig `yaml:"yandex"`
}

type HTTPConfig struct {
	Port int `yaml:"port"`
}

type TuyaConfig struct {
	Endpoint     string `yaml:"endpoint"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type YandexConfig struct {
	UserID      string      `yaml:"user_id"`
	BearerToken string      `yaml:"bearer_token"`
	PathPrefix  string      `yaml:"path_prefix"`
	OAuth       OAuthConfig `yaml:"oauth"`
}

type OAuthConfig struct {
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

	cfg.normalize()
	cfg.applyDefaults()

	return cfg, nil
}

func (cfg *Config) normalize() {
	cfg.Tuya.normalize()
	cfg.Yandex.normalize()
}

func (cfg *Config) applyDefaults() {
	cfg.Tuya.applyDefaults()
	cfg.Yandex.applyDefaults()
}

func (cfg *TuyaConfig) normalize() {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
}

func (cfg *TuyaConfig) applyDefaults() {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultTuyaEndpoint
	}
}

func (cfg *YandexConfig) normalize() {
	cfg.UserID = strings.TrimSpace(cfg.UserID)
	cfg.BearerToken = strings.TrimSpace(cfg.BearerToken)
	cfg.PathPrefix = strings.TrimSpace(cfg.PathPrefix)
	cfg.OAuth.normalize()
}

func (cfg *YandexConfig) applyDefaults() {
	if cfg.PathPrefix == "" {
		cfg.PathPrefix = DefaultYandexPathPrefix
	}
}

func (cfg *OAuthConfig) normalize() {
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
}

func (cfg HTTPConfig) Validate() error {
	if cfg.Port < 0 || cfg.Port > 65535 {
		return errors.New("http.port must be between 0 and 65535")
	}

	return nil
}

func (cfg TuyaConfig) Validate() error {
	if strings.TrimSpace(cfg.ClientID) == "" {
		return errors.New("tuya.client_id is required")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return errors.New("tuya.client_secret is required")
	}

	return nil
}

func (cfg YandexConfig) Validate() error {
	if strings.TrimSpace(cfg.UserID) == "" {
		return errors.New("yandex.user_id is required")
	}
	if strings.TrimSpace(cfg.BearerToken) == "" {
		return errors.New("yandex.bearer_token is required")
	}

	pathPrefix := strings.TrimSpace(cfg.PathPrefix)
	if pathPrefix == "" {
		return errors.New("yandex.path_prefix is required")
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		return errors.New("yandex.path_prefix must start with /")
	}
	if len(pathPrefix) > 1 && strings.HasSuffix(pathPrefix, "/") {
		return errors.New("yandex.path_prefix must not end with /")
	}

	if err := cfg.OAuth.Validate(); err != nil {
		return err
	}

	return nil
}

func (cfg OAuthConfig) Validate() error {
	if strings.TrimSpace(cfg.ClientID) == "" {
		return errors.New("yandex.oauth.client_id is required")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return errors.New("yandex.oauth.client_secret is required")
	}

	return nil
}
