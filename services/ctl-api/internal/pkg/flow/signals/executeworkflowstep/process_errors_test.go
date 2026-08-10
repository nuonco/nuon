package executeworkflowstep

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestTargetSupportsCompositeErrorHints(t *testing.T) {
	tests := []struct {
		targetType app.WorkflowStepTargetType
		want       bool
	}{
		{app.WorkflowStepTargetTypeInstallDeploy, true},
		{app.WorkflowStepTargetTypeInstallDeploys, true},
		{app.WorkflowStepTargetTypeInstallSandboxRun, true},
		{app.WorkflowStepTargetTypeInstallSandboxRuns, true},
		{app.WorkflowStepTargetTypeInstallActionWorkflowRun, false},
		{app.WorkflowStepTargetTypeInstallActionWorkflowRuns, false},
	}

	for _, test := range tests {
		t.Run(string(test.targetType), func(t *testing.T) {
			if got := targetSupportsCompositeErrorHints(string(test.targetType)); got != test.want {
				t.Fatalf("targetSupportsCompositeErrorHints(%q) = %t, want %t", test.targetType, got, test.want)
			}
		})
	}
}
