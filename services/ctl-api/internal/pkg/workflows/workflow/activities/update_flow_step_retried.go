package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateFlowStepRetriedRequest struct {
	StepID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepRetried(ctx context.Context, req UpdateFlowStepRetriedRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{ID: req.StepID}).
		Updates(map[string]any{"retried": true})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to mark step as retried")
	}

	return nil
}
