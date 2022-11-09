package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func newFakeConfig() workers.Config {
	return workers.Config{
		NuonAccessRoleArn:             uuid.NewString(),
		TokenSecretNamespace:          "default",
		OrgServerRootDomain:           "test.nuon.co",
		InstallationStateBucket:       "nuon-installations",
		InstallationStateBucketRegion: "us-west-2",
		SandboxBucket:                 "nuon-sandboxes",
	}
}

func getFakeDeprovisionRequest() DeprovisionRequest {
	return DeprovisionRequest{
		InstallID: uuid.New().String(),
		OrgID:     uuid.New().String(),
		AppID:     uuid.New().String(),

		SandboxSettings: struct {
			Name    string `json:"name" validate:"required"`
			Version string `json:"version" validate:"required"`
		}{
			Name:    "aws-eks",
			Version: "v0.1.1",
		},

		AwsRegion:     "us-west-2",
		AssumeRoleArn: uuid.NewString(),
	}
}

func TestDeprovision_finishWithErr(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()
	req := getFakeDeprovisionRequest()
	act := NewActivities(nil)

	errDestroyTerraform := fmt.Errorf("unable to destroy terraform")

	env.OnActivity(act.Start, mock.Anything, mock.Anything).
		Return(func(_ context.Context, sReq StartRequest) (StartResponse, error) {
			var resp StartResponse
			assert.NoError(t, sReq.validate())
			return resp, nil
		})

	env.OnActivity(act.DestroyTerraform, mock.Anything, mock.Anything).
		Return(func(_ context.Context, _ DestroyTerraformRequest) (DestroyTerraformResponse, error) {
			var resp DestroyTerraformResponse
			return resp, errDestroyTerraform
		})

	env.OnActivity(act.FinishDeprovision, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.NoError(t, fReq.validate())

			// verify that when a step fails, the error handler calls finish with the right params
			assert.Contains(t, fReq.ErrorMessage, errDestroyTerraform.Error())
			assert.Contains(t, fReq.ErrorStep, "destroy_terraform")
			assert.False(t, fReq.Success)

			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Deprovision, req)
}

func TestDeprovision(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()
	req := getFakeDeprovisionRequest()
	act := NewActivities(nil)

	orgID, err := shortid.ParseString(req.OrgID)
	assert.Nil(t, err)
	installID, err := shortid.ParseString(req.InstallID)
	assert.Nil(t, err)
	appID, err := shortid.ParseString(req.AppID)
	assert.Nil(t, err)

	// Mock activity implementation
	env.OnActivity(act.DestroyTerraform, mock.Anything, mock.Anything).
		Return(func(_ context.Context, dtfReq DestroyTerraformRequest) (DestroyTerraformResponse, error) {
			var resp DestroyTerraformResponse
			assert.NoError(t, dtfReq.validate())

			assert.Equal(t, orgID, dtfReq.OrgID)
			assert.Equal(t, appID, dtfReq.AppID)
			assert.Equal(t, installID, dtfReq.InstallID)
			return resp, nil
		})

	env.OnActivity(act.Start, mock.Anything, mock.Anything).
		Return(func(_ context.Context, stReq StartRequest) (StartResponse, error) {
			var resp StartResponse
			assert.NoError(t, stReq.validate())

			assert.Equal(t, orgID, stReq.OrgID)
			assert.Equal(t, appID, stReq.AppID)
			assert.Equal(t, installID, stReq.InstallID)
			return resp, nil
		})

	env.OnActivity(act.FinishDeprovision, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.NoError(t, fReq.validate())

			assert.Equal(t, orgID, fReq.OrgID)
			assert.Equal(t, appID, fReq.AppID)
			assert.Equal(t, installID, fReq.InstallID)
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Deprovision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp DeprovisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
