package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchRunStatusRequest struct {
	RunID        string `json:"run_id" validate:"required"`
	Status       string `json:"status" validate:"required,oneof=pending running success failed cancelled"`
	ErrorMessage string `json:"error_message"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateAppBranchRunStatus(ctx context.Context, req *UpdateAppBranchRunStatusRequest) (*app.AppBranchRun, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	var run app.AppBranchRun
	if err := a.db.WithContext(ctx).First(&run, "id = ?", req.RunID).Error; err != nil {
		return nil, errors.Wrap(err, "app branch run not found")
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}

	// Set timestamps based on status
	now := time.Now()
	switch req.Status {
	case "running":
		if run.StartedAt == nil {
			updates["started_at"] = now
		}
	case "success", "failed", "cancelled":
		if run.CompletedAt == nil {
			updates["completed_at"] = now
		}
	}

	// Set error message if provided
	if req.ErrorMessage != "" {
		updates["error_message"] = req.ErrorMessage
	}

	if err := a.db.WithContext(ctx).Model(&run).Updates(updates).Error; err != nil {
		return nil, errors.Wrap(err, "unable to update app branch run status")
	}

	// Reload to get updated values
	if err := a.db.WithContext(ctx).First(&run, "id = ?", req.RunID).Error; err != nil {
		return nil, errors.Wrap(err, "unable to reload app branch run")
	}

	return &run, nil
}
