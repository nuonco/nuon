package plan

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-generics"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/proto"
)

func getCreatePlanRequest(t *testing.T) *planv1.CreatePlanRequest {
	ids, err := shortid.ParseStrings(uuid.NewString(), uuid.NewString(), uuid.NewString())
	assert.NoError(t, err)
	req := generics.GetFakeObj[*planv1.CreatePlanRequest]()
	req.OrgId = ids[0]
	req.AppId = ids[1]
	req.DeploymentId = ids[2]

	return req
}

func TestCreatePlan(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := getCreatePlanRequest(t)
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

	// register activities
	a := NewActivities()
	env := testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(a.CreatePlanAct, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *planactivitiesv1.CreatePlanRequest) (*planactivitiesv1.CreatePlanResponse, error) {
			resp := &planactivitiesv1.CreatePlanResponse{}

			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.Metadata.OrgShortId)
			assert.Equal(t, req.AppId, r.Metadata.AppShortId)
			assert.Equal(t, req.DeploymentId, r.Metadata.DeploymentShortId)

			resp.Plan = planRef
			return resp, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.CreatePlan, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &planv1.CreatePlanResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.True(t, proto.Equal(planRef, resp.Plan))
}
