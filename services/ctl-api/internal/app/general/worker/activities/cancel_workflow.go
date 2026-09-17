package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type CancelWorkflowOutcome string

const (
	CancelWorkflowOutcomeCancelled CancelWorkflowOutcome = "cancelled"
	CancelWorkflowOutcomeSkipped   CancelWorkflowOutcome = "skipped"
)

type CancelWorkflowRequest struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

type CancelWorkflowResponse struct {
	Outcome CancelWorkflowOutcome `json:"outcome"`
}

// @temporal-gen-v2 activity
// @by-field WorkflowID
// @schedule-to-close-timeout 120s
// @start-to-close-timeout 60s
func (a *Activities) CancelWorkflow(ctx context.Context, req CancelWorkflowRequest) (*CancelWorkflowResponse, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("workflow_id", req.WorkflowID))

	var wf app.Workflow
	if err := a.db.WithContext(ctx).First(&wf, "id = ?", req.WorkflowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Warn("workflow not found for cancel, skipping")
			return &CancelWorkflowResponse{Outcome: CancelWorkflowOutcomeSkipped}, nil
		}
		return nil, fmt.Errorf("unable to get workflow: %w", err)
	}

	if !isCancelableStatus(wf.Status.Status) {
		return &CancelWorkflowResponse{Outcome: CancelWorkflowOutcomeSkipped}, nil
	}

	if wf.Status.Status == app.StatusPending {
		if err := a.cancelWorkflowInDB(ctx, &wf); err != nil {
			return nil, err
		}
		return &CancelWorkflowResponse{Outcome: CancelWorkflowOutcomeCancelled}, nil
	}

	if _, err := a.flowsClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: wf.ID,
	}); err != nil {
		// orphaned workflows have no live execute-flow signal left to cancel
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if dbErr := a.cancelWorkflowInDB(ctx, &wf); dbErr != nil {
				return nil, dbErr
			}
			return &CancelWorkflowResponse{Outcome: CancelWorkflowOutcomeCancelled}, nil
		}
		return nil, fmt.Errorf("unable to cancel workflow via queues: %w", err)
	}

	return &CancelWorkflowResponse{Outcome: CancelWorkflowOutcomeCancelled}, nil
}

func isCancelableStatus(status app.Status) bool {
	switch status {
	case app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
		app.StatusFailedPendingRetry:
		return true
	default:
		return false
	}
}

func (a *Activities) cancelWorkflowInDB(ctx context.Context, wf *app.Workflow) error {
	wf.Status = app.NewCompositeStatus(ctx, app.StatusCancelled)
	wf.FinishedAt = time.Now()
	if err := a.db.WithContext(ctx).Save(wf).Error; err != nil {
		return fmt.Errorf("unable to cancel workflow in db: %w", err)
	}
	return nil
}
