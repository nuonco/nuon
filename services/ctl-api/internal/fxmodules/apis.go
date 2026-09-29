package fxmodules

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/poolmetrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
)

var PublicAPIModule = fx.Module("public-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewPublicAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

var InternalAPIModule = fx.Module("internal-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewInternalAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

var RunnerAPIModule = fx.Module("runner-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewRunnerAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

var AuthAPIModule = fx.Module("auth-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewAuthAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

var AdminDashboardAPIModule = fx.Module("admin-dashboard-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewAdminDashboardAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

var SlackAPIModule = fx.Module("slack-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewSlackAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

var AllAPIsModule = fx.Module("all-apis",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(poolmetrics.New),
	fx.Provide(api.AsAPI(api.NewPublicAPI)),
	fx.Provide(api.AsAPI(api.NewRunnerAPI)),
	fx.Provide(api.AsAPI(api.NewInternalAPI)),
	fx.Provide(api.AsAPI(api.NewAuthAPI)),
	fx.Provide(api.AsAPI(api.NewAdminDashboardAPI)),
	fx.Provide(api.AsAPI(api.NewSlackAPI)),
	fx.Invoke(telemetry.StartRuntimeMetrics),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)
