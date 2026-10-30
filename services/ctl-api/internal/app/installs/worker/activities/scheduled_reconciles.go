package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallScheduledReconcileRequest struct {
	ScheduledReconcileID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ScheduledReconcileID
// @start-to-close-timeout 1m
func (a *Activities) GetInstallScheduledReconcile(ctx context.Context, req GetInstallScheduledReconcileRequest) (*app.InstallScheduledReconcile, error) {
	var schedule app.InstallScheduledReconcile
	res := a.db.WithContext(ctx).First(&schedule, "id = ?", req.ScheduledReconcileID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "get install scheduled reconcile")
	}
	return &schedule, nil
}

type MarkInstallScheduledReconcileStatusRequest struct {
	ScheduledReconcileID string                              `validate:"required"`
	Status               app.InstallScheduledReconcileStatus `validate:"required"`
	StatusReason         string
	WorkflowID           *string
}

// @temporal-gen-v2 activity
// @by-field ScheduledReconcileID
// @start-to-close-timeout 1m
func (a *Activities) MarkInstallScheduledReconcileStatus(ctx context.Context, req MarkInstallScheduledReconcileStatusRequest) error {
	updates := map[string]any{
		"status":        req.Status,
		"status_reason": req.StatusReason,
	}
	if req.WorkflowID != nil {
		updates["workflow_id"] = *req.WorkflowID
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallScheduledReconcile{}).
		Where(app.InstallScheduledReconcile{ID: req.ScheduledReconcileID}).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update scheduled reconcile status: %w", res.Error)
	}
	return nil
}

type FireInstallScheduledReconcileRequest struct {
	InstallID            string `validate:"required"`
	ScheduledReconcileID string `validate:"required"`
	PlanOnly             bool
}

type FireInstallScheduledReconcileResult struct {
	Status     app.InstallScheduledReconcileStatus `json:"status"`
	Reason     string                              `json:"reason,omitempty"`
	WorkflowID string                              `json:"workflow_id,omitempty"`
}

// FireInstallScheduledReconcile dispatches the reconcile-app-version workflow
// for a scheduled reconcile and maps the outcome to a schedule status. It is
// idempotent: if the schedule has already fired (status fired with a
// workflow_id), it returns that existing result instead of starting a second
// reconcile workflow, so a workflow retry after a crash between firing and
// recording the status cannot double-dispatch.
//
// @temporal-gen-v2 activity
// @by-field ScheduledReconcileID
// @start-to-close-timeout 2m
func (a *Activities) FireInstallScheduledReconcile(ctx context.Context, req FireInstallScheduledReconcileRequest) (*FireInstallScheduledReconcileResult, error) {
	var schedule app.InstallScheduledReconcile
	if err := a.db.WithContext(ctx).First(&schedule, "id = ?", req.ScheduledReconcileID).Error; err != nil {
		return nil, generics.TemporalGormError(err, "get install scheduled reconcile")
	}
	if schedule.Status == app.InstallScheduledReconcileStatusFired && schedule.WorkflowID != nil {
		return &FireInstallScheduledReconcileResult{
			Status:     schedule.Status,
			Reason:     schedule.StatusReason,
			WorkflowID: *schedule.WorkflowID,
		}, nil
	}
	if schedule.Status != app.InstallScheduledReconcileStatusPending {
		return &FireInstallScheduledReconcileResult{
			Status: schedule.Status,
			Reason: schedule.StatusReason,
		}, nil
	}

	result, err := a.helpers.CreateReconcileAppVersionWorkflow(ctx, installhelpers.CreateReconcileAppVersionWorkflowInput{
		InstallID: req.InstallID,
		PlanOnly:  req.PlanOnly,
	})
	switch {
	case err == nil:
		if err := a.db.WithContext(ctx).
			Model(&app.InstallScheduledReconcile{}).
			Where(app.InstallScheduledReconcile{ID: req.ScheduledReconcileID}).
			Updates(map[string]any{
				"status":      app.InstallScheduledReconcileStatusFired,
				"workflow_id": result.WorkflowID,
			}).Error; err != nil {
			return nil, fmt.Errorf("unable to record fired scheduled reconcile: %w", err)
		}
		return &FireInstallScheduledReconcileResult{
			Status:     app.InstallScheduledReconcileStatusFired,
			WorkflowID: result.WorkflowID,
		}, nil
	case errors.Is(err, installhelpers.ErrInstallAlreadyOnPinnedConfig):
		return &FireInstallScheduledReconcileResult{
			Status: app.InstallScheduledReconcileStatusNoChanges,
			Reason: "the install is already on its pinned app config",
		}, nil
	case errors.Is(err, installhelpers.ErrInstallConfigUpdateInFlight):
		return &FireInstallScheduledReconcileResult{
			Status: app.InstallScheduledReconcileStatusFailed,
			Reason: "a reconcile or branch update was already running",
		}, nil
	default:
		return nil, fmt.Errorf("unable to fire scheduled reconcile: %w", err)
	}
}
