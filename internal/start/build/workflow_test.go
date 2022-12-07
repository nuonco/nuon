package build

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func getFakeBuildRequest() BuildRequest {
	fkr := faker.New()
	var cfg BuildRequest
	fkr.Struct().Fill(&cfg)
	return cfg
}

func Test_validateBuildRequest(t *testing.T) {
	tests := map[string]struct {
		errExpectedMsg string
		buildReq       func() BuildRequest
	}{
		"should error when org id is empty": {
			errExpectedMsg: "BuildRequest.OrgID",
			buildReq: func() BuildRequest {
				req := getFakeBuildRequest()
				req.OrgID = ""
				return req
			},
		},
		"should error when app id is empty": {
			errExpectedMsg: "BuildRequest.AppID",
			buildReq: func() BuildRequest {
				req := getFakeBuildRequest()
				req.AppID = ""
				return req
			},
		},
		"should error when deployment id is empty": {
			errExpectedMsg: "BuildRequest.DeploymentID",
			buildReq: func() BuildRequest {
				req := getFakeBuildRequest()
				req.DeploymentID = ""
				return req
			},
		},
		"should not error when properly set": {
			buildReq: func() BuildRequest {
				req := getFakeBuildRequest()
				return req
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			req := test.buildReq()
			err := req.Validate()

			if test.errExpectedMsg != "" {
				assert.ErrorContains(t, err, test.errExpectedMsg)
			}
		})
	}
}

func getFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func TestProvision(t *testing.T) {
	cfg := getFakeConfig()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities(workers.Config{})

	req := getFakeBuildRequest()

	env.OnActivity(a.UpsertWaypointApplication, mock.Anything, mock.Anything).
		Return(func(_ context.Context, uwaReq UpsertWaypointApplicationRequest) (UpsertWaypointApplicationResponse, error) {
			var resp UpsertWaypointApplicationResponse
			assert.Nil(t, uwaReq.validate())
			assert.Equal(t, req.OrgID, uwaReq.OrgID)
			assert.Contains(t, uwaReq.OrgServerAddr, req.OrgID)
			return resp, nil
		})

	env.OnActivity(a.QueueWaypointDeploymentJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, uwaReq QueueWaypointDeploymentJobRequest) (QueueWaypointDeploymentJobResponse, error) {
			resp := QueueWaypointDeploymentJobResponse{
				JobID: "waypoint-job-id-abc",
			}

			assert.NoError(t, uwaReq.validate())
			assert.Equal(t, req.OrgID, uwaReq.OrgID)
			assert.Contains(t, uwaReq.OrgServerAddr, req.OrgID)
			return resp, nil
		})

	env.OnActivity(a.PollWaypointBuildJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pwdjReq PollWaypointBuildJobRequest) (PollWaypointBuildJobResponse, error) {
			var resp PollWaypointBuildJobResponse
			assert.Nil(t, pwdjReq.validate())

			assert.Equal(t, req.OrgID, pwdjReq.OrgID)
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
			assert.Equal(t, req.OrgID, vr.OrgID)
			assert.Equal(t, "waypoint-job-id-abc", vr.JobID)
			return ValidateWaypointDeploymentJobResponse{}, nil
		})

	env.ExecuteWorkflow(wkflow.Build, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp BuildResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
