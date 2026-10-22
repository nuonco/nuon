package api

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/apisurface"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/installauth"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/runnerguard"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
)

type Params struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner

	Services      []Service                `group:"services"`
	Middlewares   []middlewares.Middleware `group:"middlewares"`
	InstallAuth   *installauth.Middleware  `optional:"true"`
	APISurface    *apisurface.Middleware   `optional:"true"`
	RunnerGuard   *runnerguard.Middleware  `optional:"true"`
	L             *zap.Logger
	Cfg           *internal.Config
	DB            *gorm.DB `name:"psql"`
	EndpointAudit *EndpointAudit
	HTTPMetrics   *metrics.HTTPMetrics
}
