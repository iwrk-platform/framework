package s3

import (
	"github.com/pkg/errors"
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
		return nil, errors.New("s3 config: " + err.Error())
	}
	return &cfg, nil
}
