package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-generics"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	runnerv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/runner/v1"
	sandboxv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/sandbox/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-sender"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/workers-installs/internal/provision/sandbox"
)

func newFakeConfig() workers.Config {
	cfg := generics.GetFakeObj[workers.Config]()

	cfg.OrgInstanceRoleTemplate = "arn:aws:123456789:iam:role/org/%[1]s/org-instance-role-%[1]s"
	cfg.OrgInstallerRoleTemplate = "arn:aws:123456789:iam:role/org/%[1]s/org-installer-role-%[1]s"
	cfg.OrgInstallationsRoleTemplate = "arn:aws:123456789:iam:role/org/%[1]s/org-installations-role-%[1]s"

	return cfg
}

func TestProvision_finishWithErr(t *testing.T) {
	cfg := newFakeConfig()
	assert.NoError(t, cfg.Validate())
	req := generics.GetFakeObj[*installsv1.ProvisionRequest]()
	assert.NoError(t, req.Validate())

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	act := NewProvisionActivities(workers.Config{}, nil)

	errChildWorkflow := fmt.Errorf("unable to complete workflow")

	env.OnActivity(act.StartWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, _ StartWorkflowRequest) (StartWorkflowResponse, error) {
			var resp StartWorkflowResponse
			return resp, nil
		})

	sWkflow := sandbox.NewWorkflow(cfg)
	env.RegisterWorkflow(sWkflow.ProvisionSandbox)
	env.OnWorkflow(sWkflow.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *sandboxv1.ProvisionSandboxRequest) (*sandboxv1.ProvisionSandboxResponse, error) {
			return &sandboxv1.ProvisionSandboxResponse{}, errChildWorkflow
		})

	env.OnActivity(act.Finish, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fr FinishRequest) (FinishResponse, error) {
			assert.Equal(t, fr.ErrorStep, "provision_sandbox")
			assert.Contains(t, fr.ErrorMessage, errChildWorkflow.Error())
			assert.False(t, fr.Success)
			return FinishResponse{}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)

	var resp *installsv1.ProvisionResponse
	require.Error(t, env.GetWorkflowResult(&resp))
}

func TestProvision(t *testing.T) {
	cfg := newFakeConfig()
	assert.NoError(t, cfg.Validate())
	req := generics.GetFakeObj[*installsv1.ProvisionRequest]()
	assert.NoError(t, req.Validate())

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewProvisionActivities(cfg, sender.NewNoopSender())
	rWkflow := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(rWkflow.ProvisionRunner)
	sWkflow := sandbox.NewWorkflow(cfg)
	env.RegisterWorkflow(sWkflow.ProvisionSandbox)

	shortIDs, err := shortid.ParseStrings(req.OrgId, req.AppId, req.InstallId)
	assert.NoError(t, err)
	orgShortID, appShortID, installShortID := shortIDs[0], shortIDs[1], shortIDs[2]

	provisionOutputs := generics.GetFakeObj[sandbox.TerraformOutputs]()

	// Mock activity implementation
	env.OnWorkflow(sWkflow.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *sandboxv1.ProvisionSandboxRequest) (*sandboxv1.ProvisionSandboxResponse, error) {
			assert.Nil(t, pr.Validate())

			assert.Equal(t, orgShortID, pr.OrgId)
			assert.Equal(t, appShortID, pr.AppId)
			assert.Equal(t, installShortID, pr.InstallId)
			assert.Equal(t, req.AccountSettings, pr.AccountSettings)
			assert.Equal(t, req.SandboxSettings, pr.SandboxSettings)

			var respOutputs map[string]string
			assert.NoError(t, mapstructure.Decode(provisionOutputs, &respOutputs))
			return &sandboxv1.ProvisionSandboxResponse{TerraformOutputs: respOutputs}, nil
		})

	env.OnActivity(a.StartWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, ssnReq StartWorkflowRequest) (StartWorkflowResponse, error) {
			var resp StartWorkflowResponse
			assert.Nil(t, ssnReq.validate())
			return resp, nil
		})

	env.OnWorkflow(rWkflow.ProvisionRunner, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
			resp := &runnerv1.ProvisionRunnerResponse{}

			assert.Nil(t, r.Validate())
			assert.Equal(t, orgShortID, r.OrgId)
			assert.Equal(t, appShortID, r.AppId)
			assert.Equal(t, installShortID, r.InstallId)
			assert.Equal(t, provisionOutputs.OdrIAMRoleArn, r.OdrIamRoleArn)
			return resp, nil
		})

	env.OnActivity(a.Finish, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.Nil(t, fReq.validate())
			assert.Equal(t, orgShortID, fReq.ProvisionRequest.OrgId)
			assert.Equal(t, appShortID, fReq.ProvisionRequest.AppId)
			assert.Equal(t, installShortID, fReq.ProvisionRequest.InstallId)
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	resp := &installsv1.ProvisionResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	respTfOutputs, err := sandbox.ParseTerraformOutputs(resp.TerraformOutputs)
	assert.NoError(t, err)
	require.Equal(t, provisionOutputs, respTfOutputs)
}
