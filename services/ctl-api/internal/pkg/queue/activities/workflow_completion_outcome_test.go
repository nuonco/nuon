package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestWorkflowCompletionOutcome(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&workflowStatusRow{}))

	activities := &Activities{db: db}
	ctx := context.Background()

	tests := map[string]struct {
		status app.CompositeStatus
		want   WorkflowCompletionOutcome
	}{
		"failed pending retry": {
			status: app.CompositeStatus{Status: app.StatusFailedPendingRetry},
			want:   WorkflowCompletionOutcome{Status: app.StatusFailedPendingRetry},
		},
		"success": {
			status: app.CompositeStatus{Status: app.StatusSuccess},
			want:   WorkflowCompletionOutcome{Status: app.StatusSuccess},
		},
		"error with description": {
			status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "step deploy-app failed",
			},
			want: WorkflowCompletionOutcome{
				Status:                 app.StatusError,
				StatusHumanDescription: "step deploy-app failed",
			},
		},
		"cancelled": {
			status: app.CompositeStatus{
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "cancelled by user",
			},
			want: WorkflowCompletionOutcome{
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "cancelled by user",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, db.Save(&workflowStatusRow{
				ID:     "wfl_1",
				Status: tt.status,
			}).Error)

			outcome, err := activities.workflowCompletionOutcome(ctx, "wfl_1")
			require.NoError(t, err)
			assert.Equal(t, tt.want, *outcome)
		})
	}

	t.Run("missing workflow row returns error", func(t *testing.T) {
		outcome, err := activities.workflowCompletionOutcome(ctx, "wfl_missing")
		require.Error(t, err)
		assert.Nil(t, outcome)
	})
}

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
