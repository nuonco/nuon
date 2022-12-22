package plan

import (
	"testing"

	"github.com/powertoolsdev/go-generics"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*planv1.PlanRequest]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(wkflow.Plan, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	resp := &planv1.PlanResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
