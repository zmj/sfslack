package server

import (
	"fmt"
)

type Config struct {
	Port          int
	SfOAuthID     string
	SfOAuthSecret string
	// slack api key?
	SlackToken string
}

func (cfg *Config) validate() error {
	if cfg.Port == 0 {
		return fmt.Errorf("Missing port")
	}
	if cfg.SfOAuthID == "" {
		return fmt.Errorf("Missing oauth id")
	}
	if cfg.SfOAuthSecret == "" {
		return fmt.Errorf("Missing oauth secret")
	}
	return nil
}
