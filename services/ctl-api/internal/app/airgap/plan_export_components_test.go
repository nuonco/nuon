package airgap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestComponentSpecs(t *testing.T) {
	installComponents := []app.InstallComponent{
		{ID: "inc-1", ComponentID: "cmp-1", Component: app.Component{Name: "api", Type: "helm_chart"}},
		{ID: "inc-2", ComponentID: "cmp-2", Component: app.Component{Name: "migrations", Type: "job"}},
	}
	connections := []app.ComponentConfigConnection{
		{ComponentID: "cmp-1", HelmComponentConfig: &app.HelmComponentConfig{ChartName: "api", Namespace: generics.NewNullString("apps")}},
		{ComponentID: "cmp-2"},
	}

	specs := componentSpecs(installComponents, connections)
	require.Len(t, specs, 2)

	require.Equal(t, "inc-1", specs[0].InstallComponentID)
	require.Equal(t, "cmp-1", specs[0].ComponentID)
	require.Equal(t, "api", specs[0].ComponentName)
	require.Equal(t, "helm_chart", specs[0].ComponentType)
	require.Equal(t, "api", specs[0].HelmReleaseName)
	require.Equal(t, "apps", specs[0].HelmNamespace)

	require.Equal(t, "migrations", specs[1].ComponentName)
	require.Empty(t, specs[1].HelmReleaseName)
	require.Empty(t, specs[1].HelmNamespace)
}
