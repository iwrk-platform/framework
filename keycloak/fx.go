package keycloak

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewModule() fx.Option {
	return fx.Module(
		"keycloak",
		fx.Provide(
			newKeycloakConfig,
			newKeycloak,
		),
		fx.Invoke(func(lc fx.Lifecycle, keycloak Client) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("keycloak")
		}),
	)
}
