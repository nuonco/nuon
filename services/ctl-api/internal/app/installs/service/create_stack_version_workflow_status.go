package service

import (
	"context"
	"fmt"
	"time"

	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// createStackVersionWorkflowStep inserts a single "Generate stack version"
// step under the create-stack-version Workflow row so the dashboard's
// workflow detail page renders steps. WorkflowStep.WorkflowStepGroupID is a
// FK to workflow_step_groups, so we have to create a parent group first.
// Returns the step ID so the status helpers can advance it.
func (s *service) createStackVersionWorkflowStep(ctx context.Context, workflowID, installID string) (string, error) {
	group := app.WorkflowStepGroup{
		WorkflowID: workflowID,
		Name:       "Generate stack version",
		Status:     app.NewCompositeStatus(ctx, app.StatusPending),
		GroupIdx:   0,
		Steps: []app.WorkflowStep{{
			Name:              "Generate stack version",
			OwnerID:           installID,
			OwnerType:         "installs",
			InstallWorkflowID: workflowID,
			Status:            app.NewCompositeStatus(ctx, app.StatusPending),
			ExecutionType:     app.WorkflowStepExecutionTypeSystem,
			Idx:               0,
		}},
	}
	if res := s.db.WithContext(ctx).Create(&group); res.Error != nil {
		return "", fmt.Errorf("create workflow step group: %w", res.Error)
	}
	if len(group.Steps) == 0 {
		return "", fmt.Errorf("workflow step group created with no step")
	}
	return group.Steps[0].ID, nil
}

// markCreateStackVersionWorkflowActive flips the create-stack-version
// Workflow row and its step from Pending → InProgress and stamps StartedAt.
// Best-effort: failures only log.
func (s *service) markCreateStackVersionWorkflowActive(workflowID, stepID string) {
	now := time.Now()
	ctx := context.Background()
	if res := s.db.WithContext(ctx).
		Model(&app.Workflow{ID: workflowID}).
		Updates(app.Workflow{
			Status:    app.NewCompositeStatus(ctx, app.StatusInProgress),
			StartedAt: now,
		}); res.Error != nil {
		s.l.Warn("unable to mark create-stack-version workflow active",
			zap.String("workflow_id", workflowID), zap.Error(res.Error))
	}
	if stepID == "" {
		return
	}
	if res := s.db.WithContext(ctx).
		Model(&app.WorkflowStep{ID: stepID}).
		Updates(app.WorkflowStep{
			Status:    app.NewCompositeStatus(ctx, app.StatusInProgress),
			StartedAt: now,
		}); res.Error != nil {
		s.l.Warn("unable to mark create-stack-version step active",
			zap.String("step_id", stepID), zap.Error(res.Error))
	}
}

// awaitCreateStackVersionWorkflow blocks on the Temporal workflow's
// completion and updates the create-stack-version Workflow row + step status
// accordingly. Runs in a goroutine kicked off from the install-creation
// handler; never blocks the HTTP request.
func (s *service) awaitCreateStackVersionWorkflow(workflowID, stepID string, run tclient.WorkflowRun) {
	if run == nil {
		return
	}
	// Bound how long we'll wait: template render + S3 upload should finish in
	// under a minute. After this, give up and leave the row as InProgress so
	// it doesn't lie about success/failure.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	status := app.StatusSuccess
	if err := run.Get(ctx, nil); err != nil {
		s.l.Warn("create-stack-version workflow failed",
			zap.String("workflow_id", workflowID), zap.Error(err))
		status = app.StatusError
	}

	now := time.Now()
	if res := s.db.WithContext(ctx).
		Model(&app.Workflow{ID: workflowID}).
		Updates(app.Workflow{
			Status:     app.NewCompositeStatus(ctx, status),
			FinishedAt: now,
		}); res.Error != nil {
		s.l.Warn("unable to mark create-stack-version workflow finished",
			zap.String("workflow_id", workflowID), zap.Error(res.Error))
	}
	if stepID == "" {
		return
	}
	if res := s.db.WithContext(ctx).
		Model(&app.WorkflowStep{ID: stepID}).
		Updates(app.WorkflowStep{
			Status:     app.NewCompositeStatus(ctx, status),
			FinishedAt: now,
		}); res.Error != nil {
		s.l.Warn("unable to mark create-stack-version step finished",
			zap.String("step_id", stepID), zap.Error(res.Error))
	}
}
