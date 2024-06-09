package temporal_worker

import (
	"github.com/pkg/errors"
	config "go.uber.org/config"
)

type Config struct {
	TaskQueue string `yaml:"taskQueue"`
}

func NewWorkerConfig(provider config.Provider) (*Config, error) {
	var cfg Config
	if err := provider.Get("worker").Populate(&cfg); err != nil {
		return nil, errors.New("worker config: " + err.Error())
	}
	return &cfg, nil
}
