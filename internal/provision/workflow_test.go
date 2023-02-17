package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/go-workflows-meta/prefix"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	runnerv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/runner/v1"
	sandboxv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/sandbox/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/workers-installs/internal/provision/sandbox"
)

func newFakeConfig() internal.Config {
	cfg := generics.GetFakeObj[internal.Config]()

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
	act := NewActivities(internal.Config{}, nil)

	errChildWorkflow := fmt.Errorf("unable to complete workflow")

	sWkflow := sandbox.NewWorkflow(cfg)
	env.RegisterWorkflow(sWkflow.ProvisionSandbox)
	env.OnWorkflow(sWkflow.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *sandboxv1.ProvisionSandboxRequest) (*sandboxv1.ProvisionSandboxResponse, error) {
			return &sandboxv1.ProvisionSandboxResponse{}, errChildWorkflow
		})

	// test out meta invocations
	env.OnActivity(act.StartProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
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

	rWkflow := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(rWkflow.ProvisionRunner)
	sWkflow := sandbox.NewWorkflow(cfg)
	env.RegisterWorkflow(sWkflow.ProvisionSandbox)

	provisionOutputs := generics.GetFakeObj[sandbox.TerraformOutputs]()

	act := NewActivities(internal.Config{}, nil)
	// Mock activity implementation
	env.OnWorkflow(sWkflow.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *sandboxv1.ProvisionSandboxRequest) (*sandboxv1.ProvisionSandboxResponse, error) {
			assert.Nil(t, pr.Validate())

			assert.Equal(t, req.OrgId, pr.OrgId)
			assert.Equal(t, req.AppId, pr.AppId)
			assert.Equal(t, req.InstallId, pr.InstallId)
			assert.Equal(t, req.AccountSettings, pr.AccountSettings)
			assert.Equal(t, req.SandboxSettings, pr.SandboxSettings)

			var respOutputs map[string]string
			assert.NoError(t, mapstructure.Decode(provisionOutputs, &respOutputs))
			return &sandboxv1.ProvisionSandboxResponse{TerraformOutputs: respOutputs}, nil
		})

	env.OnWorkflow(rWkflow.ProvisionRunner, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
			resp := &runnerv1.ProvisionRunnerResponse{}

			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, req.AppId, r.AppId)
			assert.Equal(t, req.InstallId, r.InstallId)
			assert.Equal(t, provisionOutputs.OdrIAMRoleArn, r.OdrIamRoleArn)
			return resp, nil
		})

	// test out meta invocations
	env.OnActivity(act.StartProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
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
	assert.NoError(t, respTfOutputs.Validate())
}
