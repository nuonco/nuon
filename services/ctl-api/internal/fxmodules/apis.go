package fxmodules

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
)

// PublicAPIModule provides the public-facing API server.
var PublicAPIModule = fx.Module("public-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(api.AsAPI(api.NewPublicAPI)),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

// InternalAPIModule provides the internal API server.
var InternalAPIModule = fx.Module("internal-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(api.AsAPI(api.NewInternalAPI)),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

// RunnerAPIModule provides the runner API server.
var RunnerAPIModule = fx.Module("runner-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(api.AsAPI(api.NewRunnerAPI)),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

// AuthAPIModule provides the auth API server.
var AuthAPIModule = fx.Module("auth-api",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(api.AsAPI(api.NewAuthAPI)),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)

// AllAPIsModule provides all API servers (for running all in one process).
var AllAPIsModule = fx.Module("all-apis",
	fx.Provide(api.NewEndpointAudit),
	fx.Provide(api.AsAPI(api.NewPublicAPI)),
	fx.Provide(api.AsAPI(api.NewRunnerAPI)),
	fx.Provide(api.AsAPI(api.NewInternalAPI)),
	fx.Provide(api.AsAPI(api.NewAuthAPI)),
	fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
)
