package v2

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestReleaseBuildsFromWorkflow(t *testing.T) {
	releaseID := "release-id"
	componentBuildIDs := `{"config-a":"build-a","config-b":"build-b"}`
	sandboxBuildID := "sandbox-build-id"

	builds, sandbox, err := releaseBuildsFromWorkflow(&app.Workflow{Metadata: pgtype.Hstore{
		"app_release_id":              &releaseID,
		"release_component_build_ids": &componentBuildIDs,
		"release_sandbox_build_id":    &sandboxBuildID,
	}})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"config-a": "build-a", "config-b": "build-b"}, builds)
	require.Equal(t, sandboxBuildID, sandbox)
}

func TestReleaseBuildsFromWorkflowIgnoresNonReleaseWorkflow(t *testing.T) {
	builds, sandbox, err := releaseBuildsFromWorkflow(&app.Workflow{})
	require.NoError(t, err)
	require.Nil(t, builds)
	require.Empty(t, sandbox)
}
