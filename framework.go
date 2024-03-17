package v1

import (
	"github.com/iwrk-platform/framework/config"
	"github.com/iwrk-platform/framework/context"
	"github.com/iwrk-platform/framework/logger"
	"go.uber.org/fx"
)

var StandardModules = fx.Options(
	context.NewModule(),
	logger.NewModule(),
	config.NewModule(),
)
