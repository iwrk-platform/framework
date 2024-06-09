package keycloak

import "go.uber.org/zap"

type client struct {
	address      string
	realm        string
	clientId     string
	clientSecret string
	logger       *zap.Logger
	token        *token
}

type Client interface {
	GetUser(id string) (User, error)
	CreateUser(user User) error
	UpdateUser(user User) error
	DeleteUser(id string) error
	GetUserByUsername(username string) (User, error)

	SetUserClientRoles(id string, client string, roles ...string) error
	AddUserClientRoles(id string, client string, roles ...string) error
	DeleteUserClientRoles(id string, client string, roles ...string) error
}

func newKeycloak(logger *zap.Logger, config *Config) (Client, error) {
	c := &client{
		address:      config.Address,
		realm:        config.Realm,
		clientId:     config.ClientId,
		clientSecret: config.ClientSecret,
		logger:       logger,
	}

	if err := c.authorization(); err != nil {
		c.logger.Error("fault keycloak authorization", zap.Error(err))
		return nil, err
	}

	return c, nil
}
