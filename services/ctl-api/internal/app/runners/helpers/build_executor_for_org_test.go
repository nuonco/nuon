package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildExecutorForOrg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h := &Helpers{}

	t.Run("control plane", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
		}, app.RunnerJobTypeTerraformModuleBuild)

		require.NoError(t, err)
		require.Equal(t, app.RunnerJobExecutorControlPlane, executor)
		require.Empty(t, runnerID)
	})

	t.Run("rejects docker builds", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, &app.Org{
			ID: "org00000000000000000000000",
		}, app.RunnerJobTypeDockerBuild)

		require.ErrorContains(t, err, "docker_build components are not supported by control-plane builds")
		require.Equal(t, app.RunnerJobExecutorUnknown, executor)
		require.Empty(t, runnerID)
	})

	t.Run("requires org", func(t *testing.T) {
		t.Parallel()

		executor, runnerID, err := h.BuildExecutorForOrg(ctx, nil, app.RunnerJobTypeTerraformModuleBuild)

		require.Error(t, err)
		require.Equal(t, app.RunnerJobExecutorUnknown, executor)
		require.Empty(t, runnerID)
	})
}
