package s3

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewModule() fx.Option {
	return fx.Module(
		"s3",
		fx.Provide(
			newS3Config,
			newS3,
		),
		fx.Invoke(func(lc fx.Lifecycle, s3 Storage) {
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
			return log.Named("s3")
		}),
	)
}
