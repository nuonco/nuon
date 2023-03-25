package start

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1/build/v1"
	instancesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1/instances/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	workers "github.com/powertoolsdev/mono/services/workers-deployments/internal"
	"github.com/powertoolsdev/mono/services/workers-deployments/internal/start/build"
	"github.com/powertoolsdev/mono/services/workers-deployments/internal/start/instances"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/proto"
)

func getFakeStartRequest() *deploymentsv1.StartRequest {
	obj := generics.GetFakeObj[*deploymentsv1.StartRequest]()
	obj.InstallIds = []string{uuid.NewString()}
	return obj
}

func TestProvision_planOnly(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()

	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities()

	// initialize child workflows
	bld := build.NewWorkflow(cfg)
	ins := instances.NewWorkflow(cfg)
	env.RegisterWorkflow(bld.Build)
	env.RegisterWorkflow(ins.ProvisionInstances)

	req := getFakeStartRequest()
	req.PlanOnly = true
	err := req.Validate()
	assert.NoError(t, err)
	assert.True(t, req.PlanOnly)

	// Mock activity implementation
	env.OnWorkflow(bld.Build, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, br *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
			assert.True(t, br.PlanOnly)
			return &buildv1.BuildResponse{}, nil
		})
	env.OnWorkflow(ins.ProvisionInstances, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, ir *instancesv1.ProvisionRequest) (*instancesv1.ProvisionResponse, error) {
			assert.True(t, ir.PlanOnly)
			return &instancesv1.ProvisionResponse{}, nil
		})

	env.OnActivity(act.StartStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			return &sharedv1.StartActivityResponse{}, nil
		})

	env.ExecuteWorkflow(wkflow.Start, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *deploymentsv1.StartResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}

func TestStart(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	wkflow := NewWorkflow(cfg)
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities()
	bld := build.NewWorkflow(cfg)
	ins := instances.NewWorkflow(cfg)
	env.RegisterWorkflow(bld.Build)
	env.RegisterWorkflow(ins.ProvisionInstances)

	req := getFakeStartRequest()
	err := req.Validate()
	assert.NoError(t, err)

	// Mock activity implementation
	env.OnActivity(act.StartStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.DeploymentPath(req.OrgId, req.AppId, req.Component.Id, req.DeploymentId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.DeploymentPath(req.OrgId, req.AppId, req.Component.Id, req.DeploymentId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnWorkflow(bld.Build, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, br *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
			resp := &buildv1.BuildResponse{
				PlanRef: planRef,
			}
			assert.Nil(t, br.Validate())
			assert.Equal(t, req.OrgId, br.OrgId)
			return resp, nil
		})

	env.OnWorkflow(ins.ProvisionInstances, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *instancesv1.ProvisionRequest) (*instancesv1.ProvisionResponse, error) {
			resp := &instancesv1.ProvisionResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, req.AppId, r.AppId)
			assert.Equal(t, req.DeploymentId, r.DeploymentId)
			assert.True(t, proto.Equal(req.Component, r.Component))
			return resp, nil
		})

	env.ExecuteWorkflow(wkflow.Start, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *deploymentsv1.StartResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	assert.True(t, proto.Equal(planRef, resp.PlanRef))
	require.NotNil(t, resp)
}

func TestStartBuildOnly(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	wkflow := NewWorkflow(cfg)
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities()
	bld := build.NewWorkflow(cfg)
	ins := instances.NewWorkflow(cfg)
	env.RegisterWorkflow(bld.Build)
	env.RegisterWorkflow(ins.ProvisionInstances)

	req := getFakeStartRequest()
	req.BuildOnly = true
	err := req.Validate()
	assert.NoError(t, err)

	// Mock activity implementation
	env.OnActivity(act.StartStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			return resp, nil
		})

	env.OnActivity(act.FinishStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			return resp, nil
		})

	env.OnWorkflow(bld.Build, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, br *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
			resp := &buildv1.BuildResponse{
				PlanRef: planRef,
			}
			return resp, nil
		})

	env.OnWorkflow(ins.ProvisionInstances, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *instancesv1.ProvisionRequest) (*instancesv1.ProvisionResponse, error) {
			t.Errorf("provision instances called from build-only")
			return nil, nil
		})

	env.ExecuteWorkflow(wkflow.Start, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *deploymentsv1.StartResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	assert.True(t, proto.Equal(planRef, resp.PlanRef))
	require.NotNil(t, resp)
}
