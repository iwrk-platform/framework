package marina

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Marina struct {
	Client *marinaClient
	Config *Config
}

func NewMarina(logger *zap.Logger, cfg *Config) (*Marina, error) {
	sqlxConnection, err := sqlx.Open("mysql", fmt.Sprintf("server=%s;uid=%s;pwd=%s;database=%s", cfg.Host, cfg.User, cfg.Password, cfg.Database))
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConnections > 0 {
		sqlxConnection.SetMaxOpenConns(cfg.MaxOpenConnections)
	}
	return &Marina{
		Client: &marinaClient{
			Conn:   sqlxConnection,
			Logger: logger,
		},
		Config: cfg,
	}, nil
}
