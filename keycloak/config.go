package keycloak

import (
	"github.com/pkg/errors"
	"go.uber.org/config"
)

type Config struct {
	Address      string `env:"address"`
	ClientId     string `env:"client_id"`
	ClientSecret string `env:"client_secret"`
	Realm        string `env:"realm"`
}

func newKeycloakConfig(provider config.Provider) (*Config, error) {
	var cfg Config
	if err := provider.Get("keycloak").Populate(&cfg); err != nil {
		return nil, errors.New("keycloak config: " + err.Error())
	}
	return &cfg, nil
}
