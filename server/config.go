package server

import (
	"fmt"
)

type Config struct {
	Port        int
	OAuthID     string
	OAuthSecret string
	// slack api key?
}

func (cfg *Config) validate() error {
	if cfg.Port == 0 {
		return fmt.Errorf("Missing port")
	}
	if cfg.OAuthID == "" {
		return fmt.Errorf("Missing oauth id")
	}
	if cfg.OAuthSecret == "" {
		return fmt.Errorf("Missing oauth secret")
	}
	return nil
}
