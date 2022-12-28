package start

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-generics"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/powertoolsdev/workers-deployments/internal/start/build"
	"github.com/powertoolsdev/workers-deployments/internal/start/plan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
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

	act := NewActivities(cfg)
	bld := build.NewWorkflow(cfg)
	pln := plan.NewWorkflow(cfg)
	env.RegisterWorkflow(bld.Build)
	env.RegisterWorkflow(pln.Plan)

	req := getFakeStartRequest()
	req.PlanOnly = true
	err := req.Validate()
	assert.NoError(t, err)
	assert.True(t, req.PlanOnly)

	// Mock activity implementation
	env.OnActivity(act.ProvisionInstance, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionInstanceRequest) (ProvisionInstanceResponse, error) {
			assert.Falsef(t, true, "ProvisionInstance was executed during plan only")
			return ProvisionInstanceResponse{}, nil
		})

	env.OnWorkflow(bld.Build, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, br *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
			assert.Falsef(t, true, "Build was executed during plan only")
			return &buildv1.BuildResponse{}, nil
		})

	env.OnActivity(act.StartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r StartRequest) (StartResponse, error) {
			return StartResponse{}, nil
		})

	env.OnWorkflow(pln.Plan, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *planv1.PlanRequest) (*planv1.PlanResponse, error) {
			resp := &planv1.PlanResponse{}
			return resp, nil
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

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities(cfg)
	bld := build.NewWorkflow(cfg)
	pln := plan.NewWorkflow(cfg)
	env.RegisterWorkflow(bld.Build)
	env.RegisterWorkflow(pln.Plan)

	req := getFakeStartRequest()
	err := req.Validate()
	assert.NoError(t, err)

	orgShortID, err := shortid.ParseString(req.OrgId)
	assert.NoError(t, err)
	appShortID, err := shortid.ParseString(req.AppId)
	assert.NoError(t, err)
	deploymentShortID, err := shortid.ParseString(req.DeploymentId)
	assert.NoError(t, err)

	// Mock activity implementation
	env.OnActivity(act.ProvisionInstance, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionInstanceRequest) (ProvisionInstanceResponse, error) {
			assert.Nil(t, pr.validate())
			assert.Equal(t, orgShortID, pr.OrgID)
			assert.Equal(t, appShortID, pr.AppID)
			return ProvisionInstanceResponse{WorkflowID: uuid.NewString()}, nil
		})

	env.OnActivity(act.StartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r StartRequest) (StartResponse, error) {
			var resp StartResponse
			assert.Nil(t, r.validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.DeploymentsBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, orgShortID)
			assert.Equal(t, expectedRoleARN, r.DeploymentsBucketAssumeRoleARN)
			expectedPrefix := getS3Prefix(orgShortID, appShortID, req.Component.Name, deploymentShortID)
			assert.Equal(t, expectedPrefix, r.DeploymentsBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.Nil(t, r.validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.DeploymentsBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, orgShortID)
			assert.Equal(t, expectedRoleARN, r.DeploymentsBucketAssumeRoleARN)
			expectedPrefix := getS3Prefix(orgShortID, appShortID, req.Component.Name, deploymentShortID)
			assert.Equal(t, expectedPrefix, r.DeploymentsBucketPrefix)
			return resp, nil
		})

	env.OnWorkflow(bld.Build, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, br *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
			resp := &buildv1.BuildResponse{}
			assert.Nil(t, br.Validate())
			assert.Equal(t, orgShortID, br.OrgId)
			return resp, nil
		})

	env.OnWorkflow(pln.Plan, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *planv1.PlanRequest) (*planv1.PlanResponse, error) {
			resp := &planv1.PlanResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, orgShortID, r.OrgId)
			return resp, nil
		})

	env.ExecuteWorkflow(wkflow.Start, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *deploymentsv1.StartResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}

func Test_parseShortIDs(t *testing.T) {
	validUUID := uuid.NewString()

	tests := map[string]struct {
		idsFn       func() []string
		assertFn    func(*testing.T, []string)
		errExpected error
	}{
		"happy path": {
			idsFn: func() []string {
				return []string{validUUID}
			},
			assertFn: func(t *testing.T, ids []string) {
				assert.Equal(t, 1, len(ids))
				longID, err := shortid.ToUUID(ids[0])
				assert.NoError(t, err)
				assert.Equal(t, validUUID, longID.String())
			},
			errExpected: nil,
		},
		"error": {
			idsFn: func() []string {
				return []string{"invalid"}
			},
			errExpected: fmt.Errorf("invalid"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ids := test.idsFn()
			results, err := parseShortIDs(ids...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			test.assertFn(t, results)
		})
	}
}
