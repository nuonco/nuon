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
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

const maxTriggerEventDispatchAttempts = 5

type DispatchTriggerEventRequest struct {
	DispatchID      string `validate:"required"`
	GenerationToken string `validate:"required"`
}

type DispatchTriggerEventResponse struct {
	ResourceID string `json:"resource_id"`
	WorkflowID string `json:"workflow_id"`
}

type FinalizeTriggerEventDispatchFailureRequest struct {
	DispatchID      string `validate:"required"`
	GenerationToken string `validate:"required"`
	Error           string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) DispatchTriggerEvent(ctx context.Context, req DispatchTriggerEventRequest) (*DispatchTriggerEventResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, temporal.NewNonRetryableApplicationError("invalid dispatch trigger event request", "invalid_request", err)
	}

	dispatch, err := a.beginTriggerEventDispatch(ctx, req.DispatchID, req.GenerationToken)
	if err != nil {
		return nil, err
	}
	if dispatch.Status == app.EventDispatchStatusTriggered {
		return &DispatchTriggerEventResponse{ResourceID: dispatch.ResultResourceID, WorkflowID: dispatch.WorkflowID}, nil
	}

	ctx = cctx.SetOrgIDContext(ctx, dispatch.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, dispatch.CreatedByID)
	enabled, featureErr := a.featuresClient.FeatureEnabled(ctx, app.OrgFeatureTriggers)
	if featureErr != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to check triggers feature: %w", featureErr), true)
	}
	if !enabled {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, errors.New("triggers feature is not enabled"), false)
	}

	if dispatch.TargetType == app.TriggerTargetTypeRunbook {
		enabled, featureErr := a.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
		if featureErr != nil {
			return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to check runbooks feature: %w", featureErr), true)
		}
		if !enabled {
			return nil, a.failTriggerEventDispatch(ctx, dispatch, errors.New("runbooks feature is not enabled"), false)
		}
		if dispatch.RunbookConfigID == nil {
			return nil, a.failTriggerEventDispatch(ctx, dispatch, errors.New("runbook dispatch has no snapshotted config"), false)
		}
		triggered, err := a.runbooksHelpers.TriggerRunbookRun(ctx, runbookshelpers.TriggerRunbookRunRequest{InstallRunbookID: dispatch.TargetID, RunbookConfigID: *dispatch.RunbookConfigID, TriggeredByID: dispatch.CreatedByID, Inputs: dispatch.MappedInputs, TriggerEventDispatchID: &dispatch.ID})
		if err != nil {
			return nil, a.failTriggerEventDispatch(ctx, dispatch, err, true)
		}
		return a.completeTriggerEventDispatchResource(ctx, dispatch, triggered.Run.ID, triggered.Workflow.ID, plugins.TableName(a.db, app.InstallRunbookRun{}), triggered.QueueSignalID)
	}
	if dispatch.TargetType != app.TriggerTargetTypeAppBranchRun {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unsupported trigger target %q", dispatch.TargetType), false)
	}

	run, err := a.findTriggerBranchRun(ctx, dispatch.ID)
	if err == nil {
		return a.resumeTriggerBranchRun(ctx, dispatch, run)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to check existing app branch run: %w", err), true)
	}

	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Where(app.AppBranch{
		ID: dispatch.TargetID, AppID: dispatch.AppID, OrgID: dispatch.OrgID,
	}).First(&branch).Error; err != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to resolve target app branch: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}

	var branchConfig app.AppBranchConfig
	if err := a.db.WithContext(ctx).Where(app.AppBranchConfig{AppBranchID: branch.ID}).
		Order("config_number DESC").First(&branchConfig).Error; err != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to resolve target app branch config: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}

	var queue app.Queue
	if err := a.db.WithContext(ctx).Where(&app.Queue{
		OwnerID: branch.ID, OwnerType: plugins.TableName(a.db, app.AppBranch{}), Name: "",
	}, "owner_id", "owner_type", "name").First(&queue).Error; err != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to resolve target app branch queue: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}

	metadata := map[string]string{
		"app_id":                    branch.AppID,
		"config_id":                 branchConfig.ID,
		"config_number":             strconv.Itoa(branchConfig.ConfigNumber),
		"force":                     "true",
		"event_type":                "trigger",
		"trigger_event_dispatch_id": dispatch.ID,
		"trigger_event_id":          dispatch.TriggerEventID,
	}
	dedupeKey := "trigger-event-dispatch:" + dispatch.ID
	triggered, err := a.helpers.TriggerAppBranchRun(ctx, &appshelpers.TriggerAppBranchRunRequest{
		Run: appshelpers.CreateAppBranchRunRequest{
			AppBranchID:            branch.ID,
			AppBranchConfigID:      branchConfig.ID,
			Force:                  true,
			RunType:                app.AppBranchRunTypeManual,
			EventType:              "trigger",
			TriggerEventDispatchID: &dispatch.ID,
		},
		QueueID:   queue.ID,
		Metadata:  metadata,
		DedupeKey: &dedupeKey,
	})
	if err != nil {
		existing, existingErr := a.findTriggerBranchRun(ctx, dispatch.ID)
		if existingErr == nil {
			return a.resumeTriggerBranchRun(ctx, dispatch, existing)
		}
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to trigger app branch run: %w", err), true)
	}

	return a.completeTriggerEventDispatch(ctx, dispatch, triggered.Run, triggered.QueueSignalID)
}

func (a *Activities) resumeTriggerBranchRun(ctx context.Context, dispatch *app.EventDispatch, run *app.AppBranchRun) (*DispatchTriggerEventResponse, error) {
	var branchConfig app.AppBranchConfig
	if err := a.db.WithContext(ctx).Where(app.AppBranchConfig{ID: run.AppBranchConfigID}).First(&branchConfig).Error; err != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to resume app branch run config: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}
	var queue app.Queue
	if err := a.db.WithContext(ctx).Where(&app.Queue{
		OwnerID: run.AppBranchID, OwnerType: plugins.TableName(a.db, app.AppBranch{}), Name: "",
	}, "owner_id", "owner_type", "name").First(&queue).Error; err != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to resume app branch run queue: %w", err), !errors.Is(err, gorm.ErrRecordNotFound))
	}
	metadata := map[string]string{
		"app_id":                    dispatch.AppID,
		"config_id":                 branchConfig.ID,
		"config_number":             strconv.Itoa(branchConfig.ConfigNumber),
		"force":                     "true",
		"event_type":                "trigger",
		"trigger_event_dispatch_id": dispatch.ID,
		"trigger_event_id":          dispatch.TriggerEventID,
	}
	dedupeKey := "trigger-event-dispatch:" + dispatch.ID
	resumed, err := a.helpers.ResumeAppBranchRun(ctx, run, &appshelpers.TriggerAppBranchRunRequest{
		QueueID: queue.ID, Metadata: metadata, DedupeKey: &dedupeKey,
	})
	if err != nil {
		return nil, a.failTriggerEventDispatch(ctx, dispatch, fmt.Errorf("unable to resume app branch run: %w", err), true)
	}
	return a.completeTriggerEventDispatch(ctx, dispatch, resumed.Run, resumed.QueueSignalID)
}

func (a *Activities) beginTriggerEventDispatch(ctx context.Context, dispatchID, generationToken string) (*app.EventDispatch, error) {
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
			return temporal.NewNonRetryableApplicationError("stale trigger event dispatch generation", "stale_dispatch_generation", nil)
		}
		if dispatch.Status == app.EventDispatchStatusCancelled || dispatch.Status == app.EventDispatchStatusDeadLettered {
			return temporal.NewNonRetryableApplicationError("trigger event dispatch is not eligible", "dispatch_not_eligible", nil)
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
		return nil, fmt.Errorf("unable to begin trigger event dispatch: %w", err)
	}
	return &dispatch, nil
}

func (a *Activities) findTriggerBranchRun(ctx context.Context, dispatchID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	err := a.db.WithContext(ctx).Where(app.AppBranchRun{TriggerEventDispatchID: &dispatchID}).First(&run).Error
	return &run, err
}

func (a *Activities) completeTriggerEventDispatch(ctx context.Context, dispatch *app.EventDispatch, run *app.AppBranchRun, queueSignalID string) (*DispatchTriggerEventResponse, error) {
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
		return nil, fmt.Errorf("unable to mark trigger event dispatch triggered: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		var current app.EventDispatch
		if err := a.db.WithContext(ctx).Where(app.EventDispatch{ID: dispatch.ID}).First(&current).Error; err != nil {
			return nil, fmt.Errorf("unable to reload stale trigger event dispatch: %w", err)
		}
		if current.Status == app.EventDispatchStatusTriggered {
			return &DispatchTriggerEventResponse{ResourceID: current.ResultResourceID, WorkflowID: current.WorkflowID}, nil
		}
		return nil, temporal.NewNonRetryableApplicationError("stale trigger event dispatch attempt", "stale_dispatch_attempt", nil)
	}
	return &DispatchTriggerEventResponse{ResourceID: run.ID, WorkflowID: pointerValue(run.WorkflowID)}, nil
}

func (a *Activities) completeTriggerEventDispatchResource(ctx context.Context, dispatch *app.EventDispatch, resourceID, workflowID, resourceType, queueSignalID string) (*DispatchTriggerEventResponse, error) {
	triggeredAt := time.Now().UTC()
	updates := map[string]any{"status": app.EventDispatchStatusTriggered, "result_resource_type": resourceType, "result_resource_id": resourceID, "workflow_id": workflowID, "triggered_at": triggeredAt, "failed_at": nil, "next_attempt_at": nil, "error": ""}
	if queueSignalID != "" {
		updates["queue_signal_id"] = queueSignalID
	}
	result := a.db.WithContext(ctx).Model(&app.EventDispatch{}).Where(app.EventDispatch{ID: dispatch.ID, Status: app.EventDispatchStatusDispatching, ExecutionToken: dispatch.ExecutionToken}).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("unable to mark trigger event dispatch triggered: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		var current app.EventDispatch
		if err := a.db.WithContext(ctx).Where(app.EventDispatch{ID: dispatch.ID}).First(&current).Error; err != nil {
			return nil, fmt.Errorf("unable to reload stale trigger event dispatch: %w", err)
		}
		if current.Status == app.EventDispatchStatusTriggered {
			return &DispatchTriggerEventResponse{ResourceID: current.ResultResourceID, WorkflowID: current.WorkflowID}, nil
		}
		return nil, temporal.NewNonRetryableApplicationError("stale trigger event dispatch attempt", "stale_dispatch_attempt", nil)
	}
	return &DispatchTriggerEventResponse{ResourceID: resourceID, WorkflowID: workflowID}, nil
}

func (a *Activities) failTriggerEventDispatch(ctx context.Context, dispatch *app.EventDispatch, cause error, retryable bool) error {
	now := time.Now().UTC()
	status := app.EventDispatchStatusDeadLettered
	var nextAttemptAt *time.Time
	if retryable && dispatch.Attempts < maxTriggerEventDispatchAttempts {
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
		return errors.Join(cause, fmt.Errorf("unable to persist trigger event dispatch failure: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return temporal.NewNonRetryableApplicationError("stale trigger event dispatch attempt", "stale_dispatch_attempt", cause)
	}
	if status == app.EventDispatchStatusDeadLettered {
		return temporal.NewNonRetryableApplicationError(cause.Error(), "trigger_event_dispatch_dead_lettered", cause)
	}
	return cause
}

// @temporal-gen-v2 activity
func (a *Activities) FinalizeTriggerEventDispatchFailure(ctx context.Context, req FinalizeTriggerEventDispatchFailureRequest) error {
	if err := a.v.Struct(req); err != nil {
		return temporal.NewNonRetryableApplicationError("invalid finalize trigger event dispatch failure request", "invalid_request", err)
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
