package server

import (
	"fmt"
)

type Config struct {
	Host                   string
	Port                   int
	SfOAuthID              string
	SfOAuthSecret          string
	SlackVerificationToken string
}

func (cfg *Config) validate() error {
	// nothing for host - just use request host
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
