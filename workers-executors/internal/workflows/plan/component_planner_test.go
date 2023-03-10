package plan

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/proto"
)

func TestCreateComponentPlan(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

	compReq := generics.GetFakeObj[*planv1.Component]()
	cpReq := generics.GetFakeObj[*planv1.CreatePlanRequest]()
	cpReq.Type = planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD
	cpReq.Input = &planv1.CreatePlanRequest_Component{Component: compReq}

	// register activities
	a := NewActivities()
	env := testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(a.CreateComponentPlan, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *planactivitiesv1.CreateComponentPlan) (*planactivitiesv1.CreatePlanResponse, error) {
			resp := &planactivitiesv1.CreatePlanResponse{}

			assert.Nil(t, r.Validate())
			expectedReq, err := wkflow.getComponentPlanRequest(cpReq.Type, compReq)
			assert.NoError(t, err)
			assert.True(t, proto.Equal(expectedReq, r))

			assert.Equal(t, compReq.OrgId, r.Metadata.OrgShortId)
			assert.Equal(t, compReq.AppId, r.Metadata.AppShortId)
			assert.Equal(t, compReq.DeploymentId, r.Metadata.DeploymentShortId)

			resp.Plan = planRef
			return resp, nil
		})

	// TODO(jm): assert false here when a sandbox is passed in

	// execute workflow
	env.ExecuteWorkflow(wkflow.CreatePlan, cpReq)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &planv1.CreatePlanResponse{}
	assert.NoError(t, env.GetWorkflowResult(&resp))
	assert.NotNil(t, resp)
	assert.True(t, proto.Equal(planRef, resp.Plan))
}
