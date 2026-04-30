package fxmodules

import (
	"go.uber.org/fx"

	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	generalhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/general/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

// HelpersModule provides all domain-specific helper functions
// used across different parts of the application.
var HelpersModule = fx.Module("helpers",
	fx.Provide(accountshelpers.New),
	fx.Provide(vcshelpers.New),
	fx.Provide(actionshelpers.New),
	fx.Provide(componentshelpers.New),
	fx.Provide(orgshelpers.New),
	fx.Provide(appshelpers.New),
	fx.Provide(installshelpers.New),
	fx.Provide(runnershelpers.New),
	fx.Provide(generalhelpers.New),
)
