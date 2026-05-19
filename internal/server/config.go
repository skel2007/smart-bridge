package server

import "github.com/skel2007/smart-bridge/internal/config"

func validateConfig(cfg config.Config) error {
	if err := cfg.HTTP.Validate(); err != nil {
		return err
	}
	if err := cfg.Tuya.Validate(); err != nil {
		return err
	}
	if err := cfg.Yandex.Validate(); err != nil {
		return err
	}

	return nil
}
