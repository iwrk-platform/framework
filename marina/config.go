package marina

import (
	"fmt"
	"go.uber.org/config"
)

type Config struct {
	Host               string `yaml:"host"`
	Port               string `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database"`
	MaxOpenConnections int    `yaml:"maxOpenConnections"`
}

func NewMarinaConfig(provider config.Provider) (*Config, error) {
	var cfg Config
	if err := provider.Get("marina").Populate(&cfg); err != nil {
		return nil, fmt.Errorf("marina config: %w", err)
	}
	return &cfg, nil
}
