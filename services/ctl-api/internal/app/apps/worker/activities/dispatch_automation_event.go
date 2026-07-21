package activities

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

const maxAutomationDispatchAttempts = 5

type DispatchAutomationEventRequest struct {
	DispatchID      string `validate:"required"`
	GenerationToken string `validate:"required"`
}

type DispatchAutomationEventResponse struct {
	AppBranchRunID string `json:"app_branch_run_id"`
}

type FinalizeAutomationDispatchFailureRequest struct {
	DispatchID      string `validate:"required"`
	GenerationToken string `validate:"required"`
	Error           string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) DispatchAutomationEvent(ctx context.Context, req DispatchAutomationEventRequest) (*DispatchAutomationEventResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, temporal.NewNonRetryableApplicationError("invalid dispatch automation event request", "invalid_request", err)
	}

	dispatch, err := a.beginAutomationDispatch(ctx, req.DispatchID, req.GenerationToken)
	if err != nil {
		return nil, err
	}
	if dispatch.Status == app.EventDispatchStatusTriggered {
		return &DispatchAutomationEventResponse{AppBranchRunID: dispatch.ResultResourceID}, nil
	}

	ctx = cctx.SetOrgIDContext(ctx, dispatch.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, dispatch.CreatedByID)

	if dispatch.TargetType != app.EventAutomationTargetTypeAppBranchRun {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unsupported automation target %q", dispatch.TargetType), false)
	}

	run, err := a.findAutomationBranchRun(ctx, dispatch.ID)
	if err == nil {
		return a.resumeAutomationBranchRun(ctx, dispatch, run)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to check existing app branch run: %w", err), true)
	}

	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Where(app.AppBranch{
		ID: dispatch.TargetID, AppID: dispatch.AppID, OrgID: dispatch.OrgID,
	}).First(&branch).Error; err != nil {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to resolve target app branch: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}

	var branchConfig app.AppBranchConfig
	if err := a.db.WithContext(ctx).Where(app.AppBranchConfig{AppBranchID: branch.ID}).
		Order("config_number DESC").First(&branchConfig).Error; err != nil {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to resolve target app branch config: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}

	var queue app.Queue
	if err := a.db.WithContext(ctx).Where(&app.Queue{
		OwnerID: branch.ID, OwnerType: plugins.TableName(a.db, app.AppBranch{}), Name: "",
	}, "owner_id", "owner_type", "name").First(&queue).Error; err != nil {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to resolve target app branch queue: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}

	metadata := map[string]string{
		"app_id":                 branch.AppID,
		"config_id":              branchConfig.ID,
		"config_number":          strconv.Itoa(branchConfig.ConfigNumber),
		"force":                  "true",
		"event_type":             "automation",
		"automation_dispatch_id": dispatch.ID,
		"automation_event_id":    dispatch.EventSourceEventID,
	}
	dedupeKey := "automation-dispatch:" + dispatch.ID
	triggered, err := a.helpers.TriggerAppBranchRun(ctx, &appshelpers.TriggerAppBranchRunRequest{
		Run: appshelpers.CreateAppBranchRunRequest{
			AppBranchID:          branch.ID,
			AppBranchConfigID:    branchConfig.ID,
			Force:                true,
			RunType:              app.AppBranchRunTypeManual,
			EventType:            "automation",
			AutomationDispatchID: &dispatch.ID,
		},
		QueueID:   queue.ID,
		Metadata:  metadata,
		DedupeKey: &dedupeKey,
	})
	if err != nil {
		existing, existingErr := a.findAutomationBranchRun(ctx, dispatch.ID)
		if existingErr == nil {
			return a.resumeAutomationBranchRun(ctx, dispatch, existing)
		}
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to trigger app branch run: %w", err), true)
	}

	return a.completeAutomationDispatch(ctx, dispatch, triggered.Run, triggered.QueueSignalID)
}

func (a *Activities) resumeAutomationBranchRun(ctx context.Context, dispatch *app.EventDispatch, run *app.AppBranchRun) (*DispatchAutomationEventResponse, error) {
	var branchConfig app.AppBranchConfig
	if err := a.db.WithContext(ctx).Where(app.AppBranchConfig{ID: run.AppBranchConfigID}).First(&branchConfig).Error; err != nil {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to resume app branch run config: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}
	var queue app.Queue
	if err := a.db.WithContext(ctx).Where(&app.Queue{
		OwnerID: run.AppBranchID, OwnerType: plugins.TableName(a.db, app.AppBranch{}), Name: "",
	}, "owner_id", "owner_type", "name").First(&queue).Error; err != nil {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to resume app branch run queue: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}
	metadata := map[string]string{
		"app_id":                 dispatch.AppID,
		"config_id":              branchConfig.ID,
		"config_number":          strconv.Itoa(branchConfig.ConfigNumber),
		"force":                  "true",
		"event_type":             "automation",
		"automation_dispatch_id": dispatch.ID,
		"automation_event_id":    dispatch.EventSourceEventID,
	}
	dedupeKey := "automation-dispatch:" + dispatch.ID
	resumed, err := a.helpers.ResumeAppBranchRun(ctx, run, &appshelpers.TriggerAppBranchRunRequest{
		QueueID: queue.ID, Metadata: metadata, DedupeKey: &dedupeKey,
	})
	if err != nil {
		return nil, a.failAutomationDispatch(ctx, dispatch, fmt.Errorf("unable to resume app branch run: %w", err), true)
	}
	return a.completeAutomationDispatch(ctx, dispatch, resumed.Run, resumed.QueueSignalID)
}

func (a *Activities) beginAutomationDispatch(ctx context.Context, dispatchID, generationToken string) (*app.EventDispatch, error) {
	var dispatch app.EventDispatch
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.EventDispatch{ID: dispatchID}).First(&dispatch).Error; err != nil {
			return err
		}
		if dispatch.Status == app.EventDispatchStatusTriggered {
			return nil
		}
		if dispatch.GenerationToken != generationToken {
			return temporal.NewNonRetryableApplicationError("stale automation dispatch generation", "stale_dispatch_generation", nil)
		}
		if dispatch.Status == app.EventDispatchStatusCancelled || dispatch.Status == app.EventDispatchStatusDeadLettered {
			return temporal.NewNonRetryableApplicationError("automation dispatch is not eligible", "dispatch_not_eligible", nil)
		}
		dispatch.Attempts++
		dispatch.ExecutionToken = uuid.NewString()
		now := time.Now().UTC()
		return tx.Model(&dispatch).Updates(map[string]any{
			"status":          app.EventDispatchStatusDispatching,
			"attempts":        dispatch.Attempts,
			"execution_token": dispatch.ExecutionToken,
			"started_at":      now,
			"next_attempt_at": nil,
			"error":           "",
		}).Error
	})
	if err != nil {
		if temporal.IsApplicationError(err) {
			return nil, err
		}
		return nil, fmt.Errorf("unable to begin automation dispatch: %w", err)
	}
	return &dispatch, nil
}

func (a *Activities) findAutomationBranchRun(ctx context.Context, dispatchID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	err := a.db.WithContext(ctx).Where(app.AppBranchRun{AutomationDispatchID: &dispatchID}).First(&run).Error
	return &run, err
}

func (a *Activities) completeAutomationDispatch(ctx context.Context, dispatch *app.EventDispatch, run *app.AppBranchRun, queueSignalID string) (*DispatchAutomationEventResponse, error) {
	triggeredAt := time.Now().UTC()
	updates := map[string]any{
		"status":               app.EventDispatchStatusTriggered,
		"result_resource_type": plugins.TableName(a.db, app.AppBranchRun{}),
		"result_resource_id":   run.ID,
		"workflow_id":          run.WorkflowID,
		"triggered_at":         triggeredAt,
		"failed_at":            nil,
		"next_attempt_at":      nil,
		"error":                "",
	}
	if queueSignalID != "" {
		updates["queue_signal_id"] = queueSignalID
	}
	result := a.db.WithContext(ctx).Model(&app.EventDispatch{}).
		Where(app.EventDispatch{
			ID: dispatch.ID, Status: app.EventDispatchStatusDispatching, ExecutionToken: dispatch.ExecutionToken,
		}).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("unable to mark automation dispatch triggered: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		var current app.EventDispatch
		if err := a.db.WithContext(ctx).Where(app.EventDispatch{ID: dispatch.ID}).First(&current).Error; err != nil {
			return nil, fmt.Errorf("unable to reload stale automation dispatch: %w", err)
		}
		if current.Status == app.EventDispatchStatusTriggered {
			return &DispatchAutomationEventResponse{AppBranchRunID: current.ResultResourceID}, nil
		}
		return nil, temporal.NewNonRetryableApplicationError("stale automation dispatch attempt", "stale_dispatch_attempt", nil)
	}
	return &DispatchAutomationEventResponse{AppBranchRunID: run.ID}, nil
}

func (a *Activities) failAutomationDispatch(ctx context.Context, dispatch *app.EventDispatch, cause error, retryable bool) error {
	now := time.Now().UTC()
	status := app.EventDispatchStatusDeadLettered
	var nextAttemptAt *time.Time
	if retryable && dispatch.Attempts < maxAutomationDispatchAttempts {
		status = app.EventDispatchStatusRetryableFailed
		delay := 15 * time.Second * time.Duration(1<<uint(dispatch.Attempts-1))
		next := now.Add(delay)
		nextAttemptAt = &next
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	result := a.db.WithContext(cleanupCtx).Model(&app.EventDispatch{}).
		Where(app.EventDispatch{
			ID: dispatch.ID, Status: app.EventDispatchStatusDispatching, ExecutionToken: dispatch.ExecutionToken,
		}).Updates(map[string]any{
		"status":          status,
		"error":           cause.Error(),
		"failed_at":       now,
		"next_attempt_at": nextAttemptAt,
	})
	if result.Error != nil {
		return errors.Join(cause, fmt.Errorf("unable to persist automation dispatch failure: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return temporal.NewNonRetryableApplicationError("stale automation dispatch attempt", "stale_dispatch_attempt", cause)
	}
	if status == app.EventDispatchStatusDeadLettered {
		return temporal.NewNonRetryableApplicationError(cause.Error(), "automation_dispatch_dead_lettered", cause)
	}
	return cause
}

// @temporal-gen-v2 activity
func (a *Activities) FinalizeAutomationDispatchFailure(ctx context.Context, req FinalizeAutomationDispatchFailureRequest) error {
	if err := a.v.Struct(req); err != nil {
		return temporal.NewNonRetryableApplicationError("invalid finalize automation dispatch failure request", "invalid_request", err)
	}
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dispatch app.EventDispatch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.EventDispatch{ID: req.DispatchID}).First(&dispatch).Error; err != nil {
			return err
		}
		if dispatch.GenerationToken != req.GenerationToken {
			return nil
		}
		if dispatch.Status == app.EventDispatchStatusTriggered || dispatch.Status == app.EventDispatchStatusCancelled || dispatch.Status == app.EventDispatchStatusDeadLettered {
			return nil
		}
		failedAt := time.Now().UTC()
		return tx.Model(&dispatch).Updates(map[string]any{
			"status":          app.EventDispatchStatusDeadLettered,
			"error":           req.Error,
			"failed_at":       failedAt,
			"next_attempt_at": nil,
		}).Error
	})
}
