package api

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner

	Services    []Service    `group:"services"`
	Middlewares []Middleware `group:"middlewares"`
	L           *zap.Logger
	Cfg         *internal.Config
}
