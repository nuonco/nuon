package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type CancelInstallWorkflowInput struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CancelInstallWorkflow(ctx context.Context, input *CancelInstallWorkflowInput) error {
	var wf app.Workflow
	if err := a.db.WithContext(ctx).First(&wf, "id = ?", input.WorkflowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.l.Warn("workflow not found for cancel, skipping",
				zap.String("workflow_id", input.WorkflowID))
			return nil
		}
		return fmt.Errorf("unable to get workflow: %w", err)
	}

	if wf.Status.Status == app.StatusCancelled ||
		wf.Status.Status == app.StatusSuccess ||
		wf.Status.Status == app.StatusError {
		return nil
	}

	if wf.Status.Status == app.StatusPending {
		return a.cancelWorkflowInDB(ctx, &wf)
	}

	if _, err := a.flowsClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: wf.ID,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return a.cancelWorkflowInDB(ctx, &wf)
		}
		return fmt.Errorf("unable to cancel workflow via signal: %w", err)
	}

	return nil
}

func (a *Activities) cancelWorkflowInDB(ctx context.Context, wf *app.Workflow) error {
	wf.Status = app.NewCompositeStatus(ctx, app.StatusCancelled)
	wf.FinishedAt = time.Now()
	if err := a.db.WithContext(ctx).Save(wf).Error; err != nil {
		return fmt.Errorf("unable to cancel workflow in DB: %w", err)
	}
	return nil
}
