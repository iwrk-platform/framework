package s3

import (
	"fmt"
	"go.uber.org/config"
)

type Config struct {
	Host      string `yaml:"host"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
}

func newS3Config(provider config.Provider) (*Config, error) {
	var cfg Config
	if err := provider.Get("s3").Populate(&cfg); err != nil {
		return nil, fmt.Errorf("s3 config: %w", err)
	}
	return &cfg, nil
}
