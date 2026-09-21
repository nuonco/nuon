package awaitrunnerhealthy

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestValidationFailsWithMissingInstallID(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	sig := &Signal{Mode: ModeRequireActive}

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Validate(ctx)
	})

	require.ErrorContains(t, env.GetWorkflowError(), "install_id is required")
}
