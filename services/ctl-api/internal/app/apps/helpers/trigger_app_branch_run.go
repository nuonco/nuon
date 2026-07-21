package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type appBranchRunSignal struct {
	RunID string `json:"run_id"`
}

func (s *appBranchRunSignal) Type() signal.SignalType           { return "app-branch-run" }
func (s *appBranchRunSignal) Validate(_ workflow.Context) error { return nil }
func (s *appBranchRunSignal) Execute(_ workflow.Context) error  { return nil }

type TriggerAppBranchRunRequest struct {
	Run      CreateAppBranchRunRequest
	QueueID  string
	Metadata map[string]string
	Callback callback.Ref
}

type TriggerAppBranchRunResponse struct {
	Run           *app.AppBranchRun
	Workflow      *app.Workflow
	QueueSignalID string
}

func (h *Helpers) TriggerAppBranchRun(ctx context.Context, req *TriggerAppBranchRunRequest) (*TriggerAppBranchRunResponse, error) {
	run, err := h.CreateAppBranchRun(ctx, &req.Run)
	if err != nil {
		return nil, err
	}

	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["run_id"] = run.ID
	req.Metadata["commit_sha"] = run.CommitSHA
	wf, err := h.CreateWorkflow(ctx, run.AppBranchID, app.WorkflowTypeAppBranchesRun, req.Metadata, run.PlanOnly)
	if err != nil {
		triggerErr := fmt.Errorf("unable to create workflow: %w", err)
		return nil, errors.Join(triggerErr, h.failAppBranchRun(ctx, run, "workflow creation failed", err))
	}

	run.WorkflowID = &wf.ID
	if err := h.db.WithContext(ctx).Save(run).Error; err != nil {
		triggerErr := fmt.Errorf("unable to update run with workflow id: %w", err)
		return nil, errors.Join(triggerErr, h.failAppBranchRun(ctx, run, "workflow association failed", err))
	}

	enqueueResp, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   req.QueueID,
		OwnerID:   run.ID,
		OwnerType: plugins.TableName(h.db, app.AppBranchRun{}),
		Signal:    &appBranchRunSignal{RunID: run.ID},
		Callback:  req.Callback,
	})
	if err != nil {
		triggerErr := fmt.Errorf("unable to enqueue run signal: %w", err)
		return nil, errors.Join(triggerErr, h.failAppBranchRun(ctx, run, "signal enqueue failed", err))
	}

	return &TriggerAppBranchRunResponse{Run: run, Workflow: wf, QueueSignalID: enqueueResp.ID}, nil
}

func (h *Helpers) failAppBranchRun(ctx context.Context, run *app.AppBranchRun, message string, cause error) error {
	completedAt := time.Now()
	run.Status = "failed"
	run.ErrorMessage = fmt.Sprintf("%s: %v", message, cause)
	run.CompletedAt = &completedAt

	if err := h.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{ID: run.ID}).
		Updates(map[string]any{
			"status":        run.Status,
			"error_message": run.ErrorMessage,
			"completed_at":  completedAt,
		}).Error; err != nil {
		return fmt.Errorf("unable to mark app branch run failed: %w", err)
	}
	return nil
}
