package cmd

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/poolmetrics"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestAPIProviderGraphs(t *testing.T) {
	for _, tt := range []struct {
		name    string
		modules fx.Option
	}{
		{"all", fx.Options(fxmodules.AllServicesModule, fxmodules.AllAPIsModule, fxmodules.MCPAPIModule)},
		{"public", fx.Options(fxmodules.PublicServicesModule, fxmodules.PublicAPIModule)},
		{"runner", fx.Options(fxmodules.RunnerServicesModule, fxmodules.RunnerAPIModule)},
		{"auth", fx.Options(fxmodules.AuthServicesModule, fxmodules.AuthAPIModule)},
		{"internal", fx.Options(fxmodules.InternalServicesModule, fxmodules.InternalAPIModule)},
		{"admin", fx.Options(fxmodules.AdminDashboardServicesModule, fxmodules.AdminDashboardAPIModule)},
		{"slack", fx.Options(fxmodules.SlackServicesModule, fxmodules.SlackAPIModule)},
		{"mcp", fx.Options(fxmodules.MCPServicesModule, fx.Provide(api.NewEndpointAudit, poolmetrics.New), fxmodules.MCPAPIModule)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			providers := (&cli{}).providers()
			providers = append(providers, fx.NopLogger, fxmodules.MiddlewaresModule, tt.modules)
			require.NoError(t, fx.ValidateApp(providers...))
		})
	}
}
