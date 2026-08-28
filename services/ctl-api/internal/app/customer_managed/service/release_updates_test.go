package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestReleaseUpdateActiveUsesWorkflowStatus(t *testing.T) {
	workflowID := "workflow-id"
	tests := map[string]struct {
		update app.InstallAppConfigVersion
		active bool
	}{
		"orphaned update": {
			update: app.InstallAppConfigVersion{Status: app.CompositeStatus{Status: app.StatusPending}},
		},
		"stale update status": {
			update: app.InstallAppConfigVersion{
				WorkflowID: &workflowID,
				Workflow:   &app.Workflow{Status: app.CompositeStatus{Status: app.StatusSuccess}},
				Status:     app.CompositeStatus{Status: app.StatusPending},
			},
		},
		"running workflow": {
			update: app.InstallAppConfigVersion{
				WorkflowID: &workflowID,
				Workflow:   &app.Workflow{Status: app.CompositeStatus{Status: app.StatusInProgress}},
			},
			active: true,
		},
		"workflow awaiting retry": {
			update: app.InstallAppConfigVersion{
				WorkflowID: &workflowID,
				Workflow: &app.Workflow{Status: app.CompositeStatus{
					Status: app.StatusError, Metadata: map[string]any{"awaiting_retry": true},
				}},
			},
			active: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.active, releaseUpdateActive(test.update))
		})
	}
}
