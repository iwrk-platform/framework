package marina

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewModule() fx.Option {
	return fx.Module(
		"marina",
		fx.Provide(
			NewMarinaConfig,
			NewMarina,
		),
		fx.Invoke(func(lc fx.Lifecycle, m *Marina) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return m.Client.Close()
				},
			})
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("postgres")
		}),
	)
}
