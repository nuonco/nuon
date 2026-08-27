package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
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
	Run       CreateAppBranchRunRequest
	QueueID   string
	Metadata  map[string]string
	Callback  callback.Ref
	DedupeKey *string

	// ApprovalOption gates the run's plan steps. Empty means prompt.
	ApprovalOption app.InstallApprovalOption
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
	return h.ResumeAppBranchRun(ctx, run, req)
}

func (h *Helpers) ResumeAppBranchRun(ctx context.Context, run *app.AppBranchRun, req *TriggerAppBranchRunRequest) (*TriggerAppBranchRunResponse, error) {
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["run_id"] = run.ID
	req.Metadata["app_branch_id"] = run.AppBranchID
	req.Metadata["commit_sha"] = run.HeadSHA
	if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
		req.Metadata["commit_sha"] = run.VCSConnectionCommit.SHA
	}
	approvalOption := req.ApprovalOption
	if approvalOption == "" {
		approvalOption = app.InstallApprovalOptionPrompt
	}
	var wf app.Workflow
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked app.AppBranchRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.AppBranchRun{ID: run.ID}).First(&locked).Error; err != nil {
			return err
		}
		if locked.WorkflowID == nil {
			created, err := h.createWorkflowWithDB(ctx, tx, locked.AppBranchID, plugins.TableName(h.db, app.AppBranch{}), app.WorkflowTypeAppBranchesRun, req.Metadata, locked.PlanOnly, approvalOption, "")
			if err != nil {
				return fmt.Errorf("unable to create workflow: %w", err)
			}
			wf = *created
			locked.WorkflowID = &wf.ID
			if err := tx.Model(&locked).Update("workflow_id", wf.ID).Error; err != nil {
				return fmt.Errorf("unable to associate workflow: %w", err)
			}
		} else if err := tx.Where(app.Workflow{ID: *locked.WorkflowID}).First(&wf).Error; err != nil {
			return fmt.Errorf("unable to load workflow: %w", err)
		}
		*run = locked
		return nil
	})
	if err != nil {
		triggerErr := fmt.Errorf("unable to prepare workflow: %w", err)
		return nil, errors.Join(triggerErr, h.failAppBranchRun(ctx, run, "workflow preparation failed", err))
	}

	enqueueReq := &queueclient.EnqueueSignalRequest{
		QueueID:   req.QueueID,
		OwnerID:   run.ID,
		OwnerType: plugins.TableName(h.db, app.AppBranchRun{}),
		Signal:    &appBranchRunSignal{RunID: run.ID},
		Callback:  req.Callback,
		DedupeKey: req.DedupeKey,
	}
	var enqueueResp *queue.EnqueueResponse
	if req.DedupeKey != nil && *req.DedupeKey != "" {
		err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var locked app.AppBranchRun
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.AppBranchRun{ID: run.ID}).First(&locked).Error; err != nil {
				return err
			}
			var existing app.QueueSignal
			err := tx.Where(app.QueueSignal{QueueID: req.QueueID, DedupeKey: req.DedupeKey}).First(&existing).Error
			if err == nil {
				enqueueResp = &queue.EnqueueResponse{ID: existing.ID, WorkflowID: existing.Workflow.ID}
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if locked.Status == "failed" {
				if err := tx.Model(&locked).Updates(map[string]any{
					"status": "pending", "error_message": "", "completed_at": nil,
				}).Error; err != nil {
					return err
				}
			}
			var enqueueErr error
			enqueueResp, enqueueErr = h.queueClient.EnqueueSignalInTransaction(ctx, tx, enqueueReq)
			return enqueueErr
		})
		if err == nil {
			if _, wakeErr := h.queueClient.EnqueueSignal(ctx, enqueueReq); wakeErr != nil {
				h.l.Warn("unable to wake durable app branch run signal", zap.String("run_id", run.ID), zap.Error(wakeErr))
			}
		}
	} else {
		enqueueResp, err = h.queueClient.EnqueueSignal(ctx, enqueueReq)
	}
	if err != nil {
		triggerErr := fmt.Errorf("unable to enqueue run signal: %w", err)
		return nil, errors.Join(triggerErr, h.failAppBranchRun(ctx, run, "signal enqueue failed", err))
	}

	return &TriggerAppBranchRunResponse{Run: run, Workflow: &wf, QueueSignalID: enqueueResp.ID}, nil
}

func (h *Helpers) failAppBranchRun(ctx context.Context, run *app.AppBranchRun, message string, cause error) error {
	completedAt := time.Now()
	run.Status = "failed"
	run.ErrorMessage = fmt.Sprintf("%s: %v", message, cause)
	run.CompletedAt = &completedAt

	if err := h.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{ID: run.ID, Status: "pending"}).
		Updates(map[string]any{
			"status":        run.Status,
			"error_message": run.ErrorMessage,
			"completed_at":  completedAt,
		}).Error; err != nil {
		return fmt.Errorf("unable to mark app branch run failed: %w", err)
	}
	return nil
}
