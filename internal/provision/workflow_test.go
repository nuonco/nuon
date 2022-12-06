package provision

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	workers "github.com/powertoolsdev/workers-instances/internal"
)

func getFakeProvisionRequest() ProvisionRequest {
	fkr := faker.New()
	var req ProvisionRequest
	fkr.Struct().Fill(&req)
	return req
}

func Test_validateProvisionRequest(t *testing.T) {
	tests := map[string]struct {
		errExpectedMsg string
		buildReq       func() ProvisionRequest
	}{
		"should error when org id is empty": {
			errExpectedMsg: "ProvisionRequest.OrgID",
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.OrgID = ""
				return req
			},
		},
		"should error when app id is empty": {
			errExpectedMsg: "ProvisionRequest.AppID",
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.AppID = ""
				return req
			},
		},
		"should error when deployment id is empty": {
			errExpectedMsg: "ProvisionRequest.DeploymentID",
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.DeploymentID = ""
				return req
			},
		},
		"should error when no install id provided": {
			errExpectedMsg: "ProvisionRequest.InstallID",
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.InstallID = ""
				return req
			},
		},
		"should not error when properly set": {
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				return req
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			req := test.buildReq()
			err := req.validate()

			if test.errExpectedMsg != "" {
				assert.ErrorContains(t, err, test.errExpectedMsg)
			}
		})
	}
}

func TestProvision(t *testing.T) {
	wf := NewWorkflow(workers.Config{
		Bucket:                       "nuon-installations",
		WaypointTokenSecretNamespace: "default",
		WaypointServerRootDomain:     "test.nuon.co",
	})
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities()

	req := getFakeProvisionRequest()

	// Mock activity implementation
	env.OnActivity(a.GenerateWaypointConfig, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr GenerateWaypointConfigRequest) (GenerateWaypointConfigResponse, error) {
			assert.Nil(t, pr.validate())
			assert.Equal(t, req.OrgID, pr.OrgID)
			assert.Equal(t, req.AppID, pr.AppID)
			return GenerateWaypointConfigResponse{}, nil
		})

	// Mock activity implementation
	env.OnActivity(a.UpsertWaypointApplication, mock.Anything, mock.Anything).
		Return(func(_ context.Context, uwaReq UpsertWaypointApplicationRequest) (UpsertWaypointApplicationResponse, error) {
			assert.Nil(t, uwaReq.validate())

			assert.Equal(t, req.OrgID, uwaReq.OrgID)
			assert.Equal(t, req.InstallID, uwaReq.InstallID)
			assert.Equal(t, req.AppID, uwaReq.AppID)
			assert.Equal(t, req.DeploymentID, uwaReq.DeploymentID)

			return UpsertWaypointApplicationResponse{}, nil
		})

	env.OnActivity(a.QueueWaypointDeploymentJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, qwdjReq QueueWaypointDeploymentJobRequest) (QueueWaypointDeploymentJobResponse, error) {
			resp := QueueWaypointDeploymentJobResponse{
				JobID: "waypoint-job-id-abc",
			}
			assert.Nil(t, qwdjReq.validate())

			assert.Equal(t, req.OrgID, qwdjReq.OrgID)
			assert.Equal(t, req.InstallID, qwdjReq.InstallID)
			assert.Equal(t, req.AppID, qwdjReq.AppID)
			assert.Equal(t, req.DeploymentID, qwdjReq.DeploymentID)

			return resp, nil
		})

	env.OnActivity(a.PollWaypointDeploymentJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pwdjReq PollWaypointDeploymentJobRequest) (PollWaypointDeploymentJobResponse, error) {
			var resp PollWaypointDeploymentJobResponse
			assert.Nil(t, pwdjReq.validate())

			assert.Equal(t, req.OrgID, pwdjReq.OrgID)
			assert.Equal(t, "waypoint-job-id-abc", pwdjReq.JobID)

			return resp, nil
		})

	env.OnActivity(a.SendHostnameNotification, mock.Anything, mock.Anything).
		Return(func(_ context.Context, shnReq SendHostnameNotificationRequest) (SendHostnameNotificationResponse, error) {
			var resp SendHostnameNotificationResponse
			assert.Nil(t, shnReq.validate())
			assert.Equal(t, req.OrgID, shnReq.OrgID)

			return resp, nil
		})

	env.OnActivity(a.UploadMetadata, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pwdjReq UploadMetadataRequest) (*UploadResultResponse, error) {
			return nil, nil
		})

	env.ExecuteWorkflow(wf.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var resp ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
