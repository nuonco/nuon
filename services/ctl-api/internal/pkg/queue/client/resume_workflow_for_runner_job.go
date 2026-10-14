package client

import (
	"context"
	"strings"
	"time"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (c *Client) ResumeWorkflowForRunnerJob(ctx context.Context, jobID string) error {
	for {
		var job app.RunnerJob
		if err := c.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
			return err
		}
		var step app.WorkflowStep
		if err := c.db.WithContext(ctx).Where(app.WorkflowStep{ID: demand.StepID(&job)}).First(&step).Error; err != nil {
			return err
		}
		var flw app.Workflow
		if err := c.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.Workflow{ID: step.InstallWorkflowID}).First(&flw).Error; err != nil {
			return err
		}
		if flw.Status.Status == app.StatusCancelled || flw.Status.Metadata["cancel_requested_at"] != nil || step.Status.Status == app.StatusCancelled || step.RunnerJobID == nil || *step.RunnerJobID != jobID || (step.ResultDirective != demand.AwaitRunner && step.Status.Status != app.StatusPending) {
			return nil
		}
		var qs app.QueueSignal
		if err := c.db.WithContext(ctx).Where(app.QueueSignal{OwnerID: flw.ID, OwnerType: "install_workflows", Type: "execute-workflow"}).Order("created_at DESC").First(&qs).Error; err != nil {
			return err
		}
		if qs.Status.Status == app.StatusCancelled {
			return nil
		}
		handle, err := handler.UpdateWithStart(ctx, c.tClient, &qs, handler.UpdateWithStartOptions{
			UpdateName: "runner-job-completed", WaitForStage: tclient.WorkflowUpdateStageCompleted, Args: []any{struct {
				JobID string `json:"job_id"`
			}{jobID}},
		})
		if err == nil {
			err = handle.Get(ctx, nil)
		}
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "resident host finished") && !strings.Contains(err.Error(), "unknown update") && !strings.Contains(err.Error(), "closing workflow") {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
