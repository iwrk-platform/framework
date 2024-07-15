package vault

import (
	"context"
	vclient "github.com/hashicorp/vault-client-go"
	"go.uber.org/zap"
	"time"
)

const vaultTimeout = 10 * time.Second

type Vault struct {
	Client *vclient.Client
}

func newVault(logger *zap.Logger, config *cfg, ctx context.Context) (*Vault, error) {
	client, err := vclient.New(vclient.WithAddress(config.Host), vclient.WithRequestTimeout(vaultTimeout))
	if err != nil {
		logger.Error("vault.New", zap.Error(err))
		return nil, err
	}
	if err := client.SetToken(config.Token); err != nil {
		logger.Error("client.SetToken", zap.Error(err))
		return nil, err
	}
	return &Vault{Client: client}, nil
}
