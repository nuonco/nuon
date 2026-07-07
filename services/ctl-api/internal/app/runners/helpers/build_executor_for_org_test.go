package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
)

func TestBuildExecutorForOrg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h := &Helpers{}

	t.Run("org runner", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
			Features: types.StringBoolMap{
				string(app.OrgFeatureControlPlaneBuilds): false,
			},
			RunnerGroup: app.RunnerGroup{
				Runners: []app.Runner{{ID: "run00000000000000000000000"}},
			},
		}, app.RunnerJobTypeTerraformModuleBuild)

		require.NoError(t, err)
		require.Equal(t, app.RunnerJobExecutorOrgRunner, executor)
		require.Equal(t, "run00000000000000000000000", runnerID)
	})

	t.Run("control plane", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
			Features: types.StringBoolMap{
				string(app.OrgFeatureControlPlaneBuilds): true,
			},
		}, app.RunnerJobTypeTerraformModuleBuild)

		require.NoError(t, err)
		require.Equal(t, app.RunnerJobExecutorControlPlane, executor)
		require.Empty(t, runnerID)
	})

	t.Run("org runner allows docker builds", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
			Features: types.StringBoolMap{
				string(app.OrgFeatureControlPlaneBuilds): false,
			},
			RunnerGroup: app.RunnerGroup{
				Runners: []app.Runner{{ID: "run00000000000000000000000"}},
			},
		}, app.RunnerJobTypeDockerBuild)

		require.NoError(t, err)
		require.Equal(t, app.RunnerJobExecutorOrgRunner, executor)
		require.Equal(t, "run00000000000000000000000", runnerID)
	})

	t.Run("flag absent defaults to org runner", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID:       "org00000000000000000000000",
			Features: types.StringBoolMap{},
			RunnerGroup: app.RunnerGroup{
				Runners: []app.Runner{{ID: "run00000000000000000000000"}},
			},
		}, app.RunnerJobTypeDockerBuild)

		require.NoError(t, err)
		require.Equal(t, app.RunnerJobExecutorOrgRunner, executor)
		require.Equal(t, "run00000000000000000000000", runnerID)
	})

	t.Run("org runner requires runner", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
			Features: types.StringBoolMap{
				string(app.OrgFeatureControlPlaneBuilds): false,
			},
		}, app.RunnerJobTypeTerraformModuleBuild)

		require.Error(t, err)
		require.Equal(t, app.RunnerJobExecutorUnknown, executor)
		require.Empty(t, runnerID)
	})

	t.Run("control plane rejects docker builds", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
			Features: types.StringBoolMap{
				string(app.OrgFeatureControlPlaneBuilds): true,
			},
		}, app.RunnerJobTypeDockerBuild)

		require.ErrorContains(t, err, "docker_build components are not supported by control-plane builds")
		require.Equal(t, app.RunnerJobExecutorUnknown, executor)
		require.Empty(t, runnerID)
	})
}
