package start

import (
	"context"
	"log"
	"testing"

	faker "github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
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

func getFakeObj[T any]() T {
	var obj T
	err := faker.FakeData(&obj)
	if err != nil {
		log.Fatalf("unable to create fake obj: %s", err)
	}
	return obj
}

func getFakeStartRequest() *deploymentsv1.StartRequest {
	obj := getFakeObj[*deploymentsv1.StartRequest]()
	obj.InstallIds = []string{uuid.NewString()}
	return obj
}

func TestProvision(t *testing.T) {
	cfg := getFakeObj[workers.Config]()
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

	// Mock activity implementation
	env.OnActivity(act.ProvisionInstance, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionInstanceRequest) (ProvisionInstanceResponse, error) {
			assert.Nil(t, pr.validate())
			assert.Equal(t, orgShortID, pr.OrgID)
			assert.Equal(t, appShortID, pr.AppID)
			return ProvisionInstanceResponse{WorkflowID: uuid.NewString()}, nil
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
