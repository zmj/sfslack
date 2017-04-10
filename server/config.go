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
	SlackOAuthID           string
	SlackOAuthSecret       string
}

func (cfg *Config) validate() error {
	// nothing for host - just use request host
	if cfg.Port == 0 {
		return fmt.Errorf("Missing port")
	}
	if cfg.SfOAuthID == "" {
		return fmt.Errorf("Missing ShareFile oauth id")
	}
	if cfg.SfOAuthSecret == "" {
		return fmt.Errorf("Missing ShareFile oauth secret")
	}
	if cfg.SlackVerificationToken == "" {
		return fmt.Errorf("Missing Slack verification token")
	}
	if cfg.SlackOAuthID == "" {
		return fmt.Errorf("Missing Slack oauth id")
	}
	if cfg.SlackOAuthSecret == "" {
		return fmt.Errorf("Missing Slack oauth secret")
	}
	return nil
}
