package vault

import (
	"fmt"
	"go.uber.org/config"
)

type cfg struct {
	Host  string `yaml:"host"`
	Token string `yaml:"token"`
}

func newConfig(provider config.Provider) (*cfg, error) {
	var cfg cfg
	if err := provider.Get("vault").Populate(&cfg); err != nil {
		return nil, fmt.Errorf("vault config: %w", err)
	}
	return &cfg, nil
}
