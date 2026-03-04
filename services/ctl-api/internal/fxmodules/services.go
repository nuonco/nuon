package fxmodules

import (
	"go.uber.org/fx"

	accountsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/service"
	actionsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/service"
	admindashboardservice "github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service"
	appsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/service"
	authservice "github.com/nuonco/nuon/services/ctl-api/internal/app/auth/service"
	componentsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/components/service"
	generalservice "github.com/nuonco/nuon/services/ctl-api/internal/app/general/service"
	identityprovidersservice "github.com/nuonco/nuon/services/ctl-api/internal/app/identity-providers/service"
	installsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/service"
	orgsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/service"
	policyreportsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/service"
	queuesservice "github.com/nuonco/nuon/services/ctl-api/internal/app/queues/service"
	releasesservice "github.com/nuonco/nuon/services/ctl-api/internal/app/releases/service"
	runnerauthservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runner-auth/service"
	runnersservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/service"
	vcsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/service"
	"github.com/nuonco/nuon/services/ctl-api/internal/health"
	"github.com/nuonco/nuon/services/ctl-api/internal/httpbin"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/docs"
)

// domainServiceEntry defines a domain service with both its FX constructor
// and a factory for creating zero-value test instances.
type domainServiceEntry struct {
	// constructor is the FX-compatible constructor (e.g., accountsservice.New).
	constructor any
	// testFactory creates a zero-value instance for route-only tests (swagger validation).
	// Services that cannot be safely instantiated with zero-value params should set this to nil.
	testFactory func(*api.EndpointAudit) api.Service
}

// domainServices is the SINGLE SOURCE OF TRUTH for all domain services that register
// swagger-annotated routes. When adding a new domain service, add it here and both
// FX registration and swagger route tests will automatically pick it up.
var domainServices = []domainServiceEntry{
	{accountsservice.New, func(ea *api.EndpointAudit) api.Service { return accountsservice.New(accountsservice.Params{}) }},
	{actionsservice.New, func(ea *api.EndpointAudit) api.Service { return actionsservice.New(actionsservice.Params{}) }},
	{appsservice.New, func(ea *api.EndpointAudit) api.Service { return appsservice.New(appsservice.Params{EndpointAudit: ea}) }},
	{componentsservice.New, func(ea *api.EndpointAudit) api.Service { return componentsservice.New(componentsservice.Params{}) }},
	{generalservice.New, func(ea *api.EndpointAudit) api.Service { return generalservice.New(generalservice.Params{}) }},
	{identityprovidersservice.New, func(ea *api.EndpointAudit) api.Service {
		return identityprovidersservice.New(identityprovidersservice.Params{})
	}},
	{installsservice.New, func(ea *api.EndpointAudit) api.Service {
		return installsservice.New(installsservice.Params{EndpointAudit: ea})
	}},
	{orgsservice.New, func(ea *api.EndpointAudit) api.Service { return orgsservice.New(orgsservice.Params{EndpointAudit: ea}) }},
	{policyreportsservice.New, func(ea *api.EndpointAudit) api.Service {
		return policyreportsservice.New(policyreportsservice.Params{EndpointAudit: ea})
	}},
	{releasesservice.New, func(ea *api.EndpointAudit) api.Service { return releasesservice.New(releasesservice.Params{}) }},
	{runnerauthservice.New, func(ea *api.EndpointAudit) api.Service { return runnerauthservice.New(runnerauthservice.Params{}) }},
	{runnersservice.New, func(ea *api.EndpointAudit) api.Service { return runnersservice.New(runnersservice.Params{}) }},
	{vcsservice.New, func(ea *api.EndpointAudit) api.Service { return vcsservice.New(vcsservice.Params{}) }},
	{queuesservice.New, func(ea *api.EndpointAudit) api.Service { return queuesservice.New(queuesservice.Params{}) }},
}

// domainServicesFxOptions builds FX provider options from domainServices.
func domainServicesFxOptions() fx.Option {
	opts := make([]fx.Option, len(domainServices))
	for i, ds := range domainServices {
		opts[i] = fx.Provide(api.AsService(ds.constructor))
	}
	return fx.Options(opts...)
}

// TestDomainServices returns all domain services that register swagger-annotated routes,
// instantiated with zero-value params for route registration testing.
func TestDomainServices(ea *api.EndpointAudit) []api.Service {
	var svcs []api.Service
	for _, ds := range domainServices {
		if ds.testFactory != nil {
			svcs = append(svcs, ds.testFactory(ea))
		}
	}
	return svcs
}

// sharedServices are services needed by public, runner, and internal APIs.
// These do NOT include authservice which has strict config requirements.
var sharedServices = fx.Options(
	// Infrastructure services (no swagger routes).
	fx.Provide(api.AsService(docs.New)),
	fx.Provide(api.AsService(health.New)),
	fx.Provide(api.AsService(httpbin.New)),
	fx.Provide(api.AsService(admindashboardservice.New)),
	// Domain services with swagger-annotated routes.
	domainServicesFxOptions(),
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

// AdminDashboardServicesModule provides services for the admin dashboard API.
var AdminDashboardServicesModule = fx.Module("admin-dashboard-services", sharedServices)

// AllServicesModule provides all services including authservice (for dev mode).
var AllServicesModule = fx.Module("all-services",
	sharedServices,
	fx.Provide(api.AsService(authservice.New)),
)

// ServicesModule is deprecated, use API-specific modules instead.
// Kept for backwards compatibility.
var ServicesModule = AllServicesModule
