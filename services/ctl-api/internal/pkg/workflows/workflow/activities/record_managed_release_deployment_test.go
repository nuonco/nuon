package activities

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRecordsReleaseDeploymentForStateChangingManagedWorkflows(t *testing.T) {
	for _, workflowType := range []app.WorkflowType{
		app.WorkflowTypeProvision,
		app.WorkflowTypeReprovision,
		app.WorkflowTypeAppBranchConfigUpdate,
		app.WorkflowTypeDeployComponents,
		app.WorkflowTypeManualDeploy,
		app.WorkflowTypeComponentEnabled,
		app.WorkflowTypeComponentDisabled,
	} {
		require.True(t, recordsReleaseDeployment(workflowType), workflowType)
	}
	require.False(t, recordsReleaseDeployment(app.WorkflowTypeActionWorkflowRun))
	require.False(t, recordsReleaseDeployment(app.WorkflowTypeDriftRun))
	require.False(t, recordsReleaseDeployment(app.WorkflowTypeDeprovision))
}

func TestSameBuildSetRequiresExactReleaseContents(t *testing.T) {
	require.True(t, sameBuildSet(map[string]string{"api": "build-a"}, map[string]string{"api": "build-a"}))
	require.False(t, sameBuildSet(map[string]string{"api": "build-a"}, map[string]string{"api": "build-b"}))
	require.False(t, sameBuildSet(map[string]string{"api": "build-a"}, map[string]string{"api": "build-a", "worker": "build-b"}))
}

func TestWorkflowReleaseID(t *testing.T) {
	releaseID := "release-a"
	require.Equal(t, releaseID, workflowReleaseID(app.Workflow{Metadata: pgtype.Hstore{"app_release_id": &releaseID}}))
	require.Empty(t, workflowReleaseID(app.Workflow{}))
}
