package vault

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewModule() fx.Option {
	return fx.Module(
		"vault",
		fx.Provide(
			newConfig,
			newVault,
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("vault")
		}),
	)
}
