package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-generics"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestDeprovision_finishWithErr(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	assert.NoError(t, cfg.Validate())

	req := generics.GetFakeObj[*installsv1.DeprovisionRequest]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
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
	cfg := generics.GetFakeObj[workers.Config]()
	assert.NoError(t, cfg.Validate())
	req := generics.GetFakeObj[*installsv1.DeprovisionRequest]()
	assert.NoError(t, req.Validate())

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	act := NewActivities(nil)

	orgID, err := shortid.ParseString(req.OrgId)
	assert.Nil(t, err)
	installID, err := shortid.ParseString(req.InstallId)
	assert.Nil(t, err)
	appID, err := shortid.ParseString(req.AppId)
	assert.Nil(t, err)

	// Mock activity implementation
	env.OnActivity(act.DestroyTerraform, mock.Anything, mock.Anything).
		Return(func(_ context.Context, dtfReq DestroyTerraformRequest) (DestroyTerraformResponse, error) {
			var resp DestroyTerraformResponse
			assert.NoError(t, dtfReq.validate())

			assert.Equal(t, orgID, dtfReq.DeprovisionRequest.OrgId)
			assert.Equal(t, appID, dtfReq.DeprovisionRequest.AppId)
			assert.Equal(t, installID, dtfReq.DeprovisionRequest.InstallId)
			return resp, nil
		})

	env.OnActivity(act.Start, mock.Anything, mock.Anything).
		Return(func(_ context.Context, stReq StartRequest) (StartResponse, error) {
			var resp StartResponse
			assert.NoError(t, stReq.validate())

			assert.Equal(t, orgID, stReq.DeprovisionRequest.OrgId)
			assert.Equal(t, appID, stReq.DeprovisionRequest.AppId)
			assert.Equal(t, installID, stReq.DeprovisionRequest.InstallId)
			return resp, nil
		})

	env.OnActivity(act.FinishDeprovision, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.NoError(t, fReq.validate())

			assert.Equal(t, orgID, fReq.DeprovisionRequest.OrgId)
			assert.Equal(t, appID, fReq.DeprovisionRequest.AppId)
			assert.Equal(t, installID, fReq.DeprovisionRequest.InstallId)
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Deprovision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *installsv1.DeprovisionRequest
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
