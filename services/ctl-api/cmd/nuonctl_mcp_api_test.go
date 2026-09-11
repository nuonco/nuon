package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func TestNuonctlMCPAPIGraph(t *testing.T) {
	c := &cli{}
	options := c.providers()
	options = append(options,
		fxmodules.NuonctlMCPServicesModule,
		fxmodules.NuonctlMCPAPIModule,
	)

	require.NoError(t, fx.ValidateApp(options...))
}
