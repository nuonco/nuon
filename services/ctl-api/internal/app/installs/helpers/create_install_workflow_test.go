package helpers

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRequiresLiveInstallRunner(t *testing.T) {
	tests := map[string]struct {
		workflowType app.WorkflowType
		metadata     map[string]string
		expected     bool
	}{
		"deprovision": {
			workflowType: app.WorkflowTypeDeprovision,
			expected:     true,
		},
		"provision": {
			workflowType: app.WorkflowTypeProvision,
			expected:     false,
		},
		"input update": {
			workflowType: app.WorkflowTypeInputUpdate,
			expected:     true,
		},
		"inputs only": {
			workflowType: app.WorkflowTypeInputUpdate,
			metadata: map[string]string{
				app.WorkflowMetadataKeyInputsOnly: "true",
			},
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if actual := requiresLiveInstallRunner(tt.workflowType, tt.metadata); actual != tt.expected {
				t.Fatalf("expected %t, got %t", tt.expected, actual)
			}
		})
	}
}
