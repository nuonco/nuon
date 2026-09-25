package activities

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestWorkflowCompletionOutcomeHumanDescription(t *testing.T) {
	tests := map[string]struct {
		outcome WorkflowCompletionOutcome
		want    string
	}{
		"writer-provided description wins": {
			outcome: WorkflowCompletionOutcome{
				Status:                 app.StatusError,
				StatusHumanDescription: "step deploy-app failed",
			},
			want: "step deploy-app failed",
		},
		"empty error description defaults": {
			outcome: WorkflowCompletionOutcome{Status: app.StatusError},
			want:    "workflow failed",
		},
		"empty cancelled description defaults": {
			outcome: WorkflowCompletionOutcome{Status: app.StatusCancelled},
			want:    "workflow cancelled",
		},
		"non-terminal statuses stay empty": {
			outcome: WorkflowCompletionOutcome{Status: app.StatusSuccess},
			want:    "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.outcome.HumanDescription())
		})
	}
}
