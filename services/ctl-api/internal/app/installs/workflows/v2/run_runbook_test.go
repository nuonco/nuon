package v2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRunbookStepWorkflow(t *testing.T) {
	workflow := &app.Workflow{}

	planOnly := runbookStepWorkflow(workflow, &app.RunbookStepConfig{
		Type:     app.RunbookStepTypeComponentDeploy,
		PlanOnly: true,
	})
	require.True(t, planOnly.PlanOnly)
	require.False(t, workflow.PlanOnly)

	nextStep := runbookStepWorkflow(workflow, &app.RunbookStepConfig{Type: app.RunbookStepTypeComponentDeploy})
	require.False(t, nextStep.PlanOnly)

	sandbox := runbookStepWorkflow(workflow, &app.RunbookStepConfig{
		Type:     app.RunbookStepTypeSandboxReprovision,
		PlanOnly: true,
	})
	require.True(t, sandbox.PlanOnly)
	require.False(t, workflow.PlanOnly)

	action := runbookStepWorkflow(workflow, &app.RunbookStepConfig{
		Type:     app.RunbookStepTypeAction,
		PlanOnly: true,
	})
	require.False(t, action.PlanOnly)
}
