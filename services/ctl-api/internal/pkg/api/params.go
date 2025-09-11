package api

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner

	Services      []Service                `group:"services"`
	Middlewares   []middlewares.Middleware `group:"middlewares"`
	L             *zap.Logger
	Cfg           *internal.Config
	DB            *gorm.DB `name:"psql"`
	EndpointAudit *EndpointAudit
}
