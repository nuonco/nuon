package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	awseks "github.com/powertoolsdev/mono/pkg/sandboxes/aws-eks"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	dnsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/dns/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/activities"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/dns"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/runner"
)

// NOTE(jm): unfortunately, the only way to register these workflows in the test env is to do it using the same exact
// signature. Given we'll be using these workflows from just about every domain, we should probably make a library to
// wrap these calls, so we don't have to maintain them everywhere like this.
func CreatePlan(workflow.Context, *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	return nil, nil
}

func ExecutePlan(workflow.Context, *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	return nil, nil
}

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
	req.AzureSettings = nil
	req.PlanOnly = false
	req.RunnerType = installsv1.RunnerType_RUNNER_TYPE_AWS_EKS
	assert.NoError(t, req.Validate())

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	act := NewActivities(nil, nil, nil)

	errChildWorkflow := fmt.Errorf("unable to complete workflow")

	env.RegisterWorkflow(CreatePlan)
	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			return &planv1.CreatePlanResponse{}, errChildWorkflow
		})

	env.RegisterWorkflow(ExecutePlan)
	env.OnWorkflow("ExecutePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			return &executev1.ExecutePlanResponse{}, errChildWorkflow
		})

	// test out meta invocations
	env.OnActivity(act.StartProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			assertedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, assertedRoleARN, r.MetadataBucketAssumeRoleArn)
			assertedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, assertedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			assertedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, assertedRoleARN, r.MetadataBucketAssumeRoleArn)
			assertedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, assertedPrefix, r.MetadataBucketPrefix)

			var wkflowReq installsv1.ProvisionRequest
			err := r.Other.UnmarshalTo(&wkflowReq)
			assert.NoError(t, err)
			assert.True(t, proto.Equal(&wkflowReq, req))

			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)

	var resp *installsv1.ProvisionResponse
	assert.Error(t, env.GetWorkflowResult(&resp))
}

func TestProvision(t *testing.T) {
	cfg := newFakeConfig()
	assert.NoError(t, cfg.Validate())
	req := generics.GetFakeObj[*installsv1.ProvisionRequest]()
	req.AzureSettings = nil
	assert.NoError(t, req.Validate())
	req.PlanOnly = false
	req.RunnerType = installsv1.RunnerType_RUNNER_TYPE_AWS_EKS
	planref := generics.GetFakeObj[*planv1.PlanRef]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	rWkflow := runner.NewWorkflow(nil, cfg)
	dnsWkflow := dns.NewWorkflow(cfg)
	env.RegisterWorkflow(rWkflow.ProvisionRunner)
	env.RegisterWorkflow(dnsWkflow.ProvisionDNS)
	env.RegisterWorkflow(CreatePlan)
	env.RegisterWorkflow(ExecutePlan)

	provisionOutputs := generics.GetFakeObj[awseks.TerraformOutputs]()
	assert.NoError(t, provisionOutputs.Validate())

	act := NewActivities(nil, nil, nil)
	// Mock activity implementation
	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			assert.Nil(t, pr.Validate())

			assert.Equal(t, req.OrgId, pr.GetSandbox().OrgId)
			assert.Equal(t, req.AppId, pr.GetSandbox().AppId)
			assert.Equal(t, req.InstallId, pr.GetSandbox().InstallId)
			assert.Equal(t, req.RunId, pr.GetSandbox().RunId)
			assert.Equal(t, req.SandboxSettings, pr.GetSandbox().SandboxSettings)
			assert.Equal(t, req.AwsSettings, pr.GetSandbox().AwsSettings)

			return &planv1.CreatePlanResponse{Plan: planref}, nil
		})

	env.OnWorkflow("ExecutePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			assert.Nil(t, pr.Validate())
			assert.Equal(t, planref, pr.Plan)
			return &executev1.ExecutePlanResponse{}, nil
		})

	env.OnWorkflow(dnsWkflow.ProvisionDNS, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *dnsv1.ProvisionDNSRequest) (*dnsv1.ProvisionDNSResponse, error) {
			resp := &dnsv1.ProvisionDNSResponse{}

			assert.Nil(t, r.Validate())
			assert.Equal(t, r.Nameservers, awseks.ToStringSlice(provisionOutputs.PublicDomain.Nameservers))
			assert.Equal(t, r.Domain, provisionOutputs.PublicDomain.Name)
			return resp, nil
		})

	env.OnWorkflow(rWkflow.ProvisionRunner, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
			resp := &runnerv1.ProvisionRunnerResponse{}

			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, req.AppId, r.AppId)
			assert.Equal(t, req.InstallId, r.InstallId)
			assert.Equal(t, provisionOutputs.Runner.DefaultIAMRoleARN, r.OdrIamRoleArn)
			return resp, nil
		})

	// test out meta invocations
	env.OnActivity(act.StartProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			assertedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, assertedRoleARN, r.MetadataBucketAssumeRoleArn)
			assertedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, assertedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			assertedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, assertedRoleARN, r.MetadataBucketAssumeRoleArn)
			assertedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, assertedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.CheckIAMRole, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r CheckIAMRoleRequest) (CheckIAMRoleResponse, error) {
			resp := CheckIAMRoleResponse{}
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	resp := &installsv1.ProvisionResponse{}
	assert.NoError(t, env.GetWorkflowResult(&resp))
	assert.NotNil(t, resp)
	// respTfOutputs, err := sandbox.ParseTerraformOutputs(resp.TerraformOutputs)
	// assert.NoError(t, err)
	// assert.NoError(t, respTfOutputs.Validate())
}

func TestProvision_plan_only(t *testing.T) {
	cfg := newFakeConfig()
	assert.NoError(t, cfg.Validate())
	req := generics.GetFakeObj[*installsv1.ProvisionRequest]()
	req.AzureSettings = nil
	assert.NoError(t, req.Validate())
	req.PlanOnly = true
	req.RunnerType = installsv1.RunnerType_RUNNER_TYPE_AWS_EKS
	planref := generics.GetFakeObj[*planv1.PlanRef]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	rWkflow := runner.NewWorkflow(nil, cfg)
	dnsWkflow := dns.NewWorkflow(cfg)
	env.RegisterWorkflow(rWkflow.ProvisionRunner)
	env.RegisterWorkflow(dnsWkflow.ProvisionDNS)
	env.RegisterWorkflow(CreatePlan)
	env.RegisterWorkflow(ExecutePlan)

	act := NewActivities(nil, nil, nil)
	sharedActs := activities.NewActivities(nil, nil)
	// Mock activity implementation
	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			assert.Nil(t, pr.Validate())

			assert.Equal(t, req.OrgId, pr.GetSandbox().OrgId)
			assert.Equal(t, req.AppId, pr.GetSandbox().AppId)
			assert.Equal(t, req.InstallId, pr.GetSandbox().InstallId)
			assert.Equal(t, req.SandboxSettings, pr.GetSandbox().SandboxSettings)
			assert.Equal(t, req.AwsSettings, pr.GetSandbox().AwsSettings)

			return &planv1.CreatePlanResponse{Plan: planref}, nil
		})

	// test out meta invocations
	env.OnActivity(act.StartProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			assertedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, assertedRoleARN, r.MetadataBucketAssumeRoleArn)
			assertedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, assertedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.InstallationsBucket, r.MetadataBucket)

			assertedRoleARN := fmt.Sprintf(cfg.OrgInstallationsRoleTemplate, req.OrgId)
			assert.Equal(t, assertedRoleARN, r.MetadataBucketAssumeRoleArn)
			assertedPrefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
			assert.Equal(t, assertedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(sharedActs.FetchSandboxOutputs, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r activities.FetchSandboxOutputsRequest) (*structpb.Struct, error) {
			return &structpb.Struct{}, nil
		})

	env.OnActivity(act.CheckIAMRole, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r CheckIAMRoleRequest) (CheckIAMRoleResponse, error) {
			resp := CheckIAMRoleResponse{}
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	resp := &installsv1.ProvisionResponse{}
	assert.NoError(t, env.GetWorkflowResult(&resp))
	assert.NotNil(t, resp)
	// respTfOutputs, err := sandbox.ParseTerraformOutputs(resp.TerraformOutputs)
	// assert.NoError(t, err)
	// assert.NoError(t, respTfOutputs.Validate())
}
