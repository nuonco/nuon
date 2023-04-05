package provision

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	req := generics.GetFakeObj[*canaryv1.ProvisionRequest]()

	// Mock activity implementations

	// execute workflow
	env.ExecuteWorkflow(wkflow.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify response
	var resp *canaryv1.ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
