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

// ServicesModule provides all API endpoint services (domain handlers).
var ServicesModule = fx.Module("services",
	fx.Provide(api.AsService(docs.New)),
	fx.Provide(api.AsService(health.New)),
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
	fx.Provide(api.AsService(httpbin.New)),
	fx.Provide(api.AsService(authservice.New)),
)
