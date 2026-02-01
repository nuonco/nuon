package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/admin"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/auth"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/chaos"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/cors"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/global"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/headers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/invites"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/org"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/pagination"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/panicker"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/patcher"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/public"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/size"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/timeout"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/tracer"
)

// MiddlewaresModule provides all HTTP middlewares used by the API services.
var MiddlewaresModule = fx.Module("middlewares",
	fx.Provide(middlewares.AsMiddleware(stderr.New)),
	fx.Provide(middlewares.AsMiddleware(global.New)),
	fx.Provide(middlewares.AsMiddleware(metrics.New)),
	fx.Provide(middlewares.AsMiddleware(metrics.NewInternal)),
	fx.Provide(middlewares.AsMiddleware(metrics.NewRunner)),
	fx.Provide(middlewares.AsMiddleware(headers.New)),
	fx.Provide(middlewares.AsMiddleware(auth.New)),
	fx.Provide(middlewares.AsMiddleware(org.New)),
	fx.Provide(middlewares.AsMiddleware(org.NewRunner)),
	fx.Provide(middlewares.AsMiddleware(public.New)),
	fx.Provide(middlewares.AsMiddleware(pagination.New)),
	fx.Provide(middlewares.AsMiddleware(cors.New)),
	fx.Provide(middlewares.AsMiddleware(config.New)),
	fx.Provide(middlewares.AsMiddleware(patcher.New)),
	fx.Provide(middlewares.AsMiddleware(invites.New)),
	fx.Provide(middlewares.AsMiddleware(admin.New)),
	fx.Provide(middlewares.AsMiddleware(log.New)),
	fx.Provide(middlewares.AsMiddleware(size.New)),
	fx.Provide(middlewares.AsMiddleware(timeout.New)),
	fx.Provide(middlewares.AsMiddleware(audit.NewPublic)),
	fx.Provide(middlewares.AsMiddleware(audit.NewInternal)),
	fx.Provide(middlewares.AsMiddleware(audit.NewRunner)),
	fx.Provide(middlewares.AsMiddleware(panicker.New)),
	fx.Provide(middlewares.AsMiddleware(tracer.New)),
	fx.Provide(middlewares.AsMiddleware(chaos.New)),
)
