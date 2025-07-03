package helpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallWorkflowStepRetry struct {
	StepID string `validate:"required"`
}

// UpdateInstallWorkflowStepRetry updates the retry status of an install workflow step.
// This makes the step non retryable for next attempts.
func (h *Helpers) UpdateInstallWorkflowStepRetry(ctx context.Context, req UpdateInstallWorkflowStepRetry) error {
	step := app.WorkflowStep{
		ID: req.StepID,
	}

	res := h.db.WithContext(ctx).
		Model(&step).
		Updates(app.WorkflowStep{
			Retried: true,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update install workflow step retryable status")
	}
	if res.RowsAffected == 0 {
		return errors.Errorf("install workflow step not found")
	}
	return nil
}
