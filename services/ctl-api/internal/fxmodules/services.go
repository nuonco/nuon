package fxmodules

import (
	"go.uber.org/fx"

	accountsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/service"
	actionsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/service"
	appsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/service"
	authservice "github.com/nuonco/nuon/services/ctl-api/internal/app/auth/service"
	componentsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/components/service"
	generalservice "github.com/nuonco/nuon/services/ctl-api/internal/app/general/service"
	installsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/service"
	orgsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/service"
	releasesservice "github.com/nuonco/nuon/services/ctl-api/internal/app/releases/service"
	runnersservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/service"
	vcsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/service"
	"github.com/nuonco/nuon/services/ctl-api/internal/health"
	"github.com/nuonco/nuon/services/ctl-api/internal/httpbin"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/docs"
)

// sharedServices are services needed by public, runner, and internal APIs.
// These do NOT include authservice which has strict config requirements.
var sharedServices = fx.Options(
	fx.Provide(api.AsService(docs.New)),
	fx.Provide(api.AsService(health.New)),
	fx.Provide(api.AsService(httpbin.New)),
	fx.Provide(api.AsService(accountsservice.New)),
	fx.Provide(api.AsService(orgsservice.New)),
	fx.Provide(api.AsService(appsservice.New)),
	fx.Provide(api.AsService(vcsservice.New)),
	fx.Provide(api.AsService(generalservice.New)),
	fx.Provide(api.AsService(installsservice.New)),
	fx.Provide(api.AsService(componentsservice.New)),
	fx.Provide(api.AsService(runnersservice.New)),
	fx.Provide(api.AsService(releasesservice.New)),
	fx.Provide(api.AsService(actionsservice.New)),
)

// PublicServicesModule provides services for the public API (excludes authservice).
var PublicServicesModule = fx.Module("public-services", sharedServices)

// RunnerServicesModule provides services for the runner API (excludes authservice).
var RunnerServicesModule = fx.Module("runner-services", sharedServices)

// InternalServicesModule provides services for the internal API (excludes authservice).
var InternalServicesModule = fx.Module("internal-services", sharedServices)

// AuthServicesModule provides services for the auth API (includes authservice).
var AuthServicesModule = fx.Module("auth-services",
	sharedServices,
	fx.Provide(api.AsService(authservice.New)),
)

// AllServicesModule provides all services including authservice (for dev mode).
var AllServicesModule = fx.Module("all-services",
	sharedServices,
	fx.Provide(api.AsService(authservice.New)),
)

// ServicesModule is deprecated, use API-specific modules instead.
// Kept for backwards compatibility.
var ServicesModule = AllServicesModule
