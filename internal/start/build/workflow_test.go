package build

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*buildv1.BuildRequest]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities(workers.Config{})

	env.OnActivity(a.UpsertWaypointApplication, mock.Anything, mock.Anything).
		Return(func(_ context.Context, uwaReq UpsertWaypointApplicationRequest) (UpsertWaypointApplicationResponse, error) {
			var resp UpsertWaypointApplicationResponse
			assert.Nil(t, uwaReq.validate())
			assert.Equal(t, req.OrgId, uwaReq.OrgID)
			assert.Contains(t, uwaReq.OrgServerAddr, req.OrgId)
			return resp, nil
		})

	env.OnActivity(a.QueueWaypointDeploymentJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, uwaReq QueueWaypointDeploymentJobRequest) (QueueWaypointDeploymentJobResponse, error) {
			resp := QueueWaypointDeploymentJobResponse{
				JobID: "waypoint-job-id-abc",
			}

			assert.NoError(t, uwaReq.validate())
			assert.Equal(t, req.OrgId, uwaReq.OrgID)
			assert.Contains(t, uwaReq.OrgServerAddr, req.OrgId)
			return resp, nil
		})

	env.OnActivity(a.PollWaypointBuildJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pwdjReq PollWaypointBuildJobRequest) (PollWaypointBuildJobResponse, error) {
			var resp PollWaypointBuildJobResponse
			assert.Nil(t, pwdjReq.validate())

			assert.Equal(t, req.OrgId, pwdjReq.OrgID)
			assert.Equal(t, "waypoint-job-id-abc", pwdjReq.JobID)

			return resp, nil
		})

	env.OnActivity(a.UploadArtifact, mock.Anything, mock.Anything).
		Return(func(_ context.Context, ua UploadArtifactRequest) (UploadArtifactResponse, error) {
			assert.Nil(t, ua.validate())
			return UploadArtifactResponse{}, nil
		})

	env.OnActivity(a.ValidateWaypointDeploymentJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, vr ValidateWaypointDeploymentJobRequest) (ValidateWaypointDeploymentJobResponse, error) {
			assert.Nil(t, vr.validate())
			assert.Equal(t, req.OrgId, vr.OrgID)
			assert.Equal(t, "waypoint-job-id-abc", vr.JobID)
			return ValidateWaypointDeploymentJobResponse{}, nil
		})

	env.ExecuteWorkflow(wkflow.Build, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	resp := &buildv1.BuildResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
