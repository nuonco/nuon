package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateFlowStepGroupResultDirectiveRequest struct {
	StepGroupID string `validate:"required"`
	Directive   string // empty string is valid (used to clear the directive)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepGroupResultDirective(ctx context.Context, req UpdateFlowStepGroupResultDirectiveRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.WorkflowStepGroup{}).
		Where(app.WorkflowStepGroup{ID: req.StepGroupID}).
		// Must use map, not struct — GORM's struct-based Updates() skips zero-value
		// fields, so Updates(app.WorkflowStepGroup{ResultDirective: ""}) would be a no-op
		// when clearing the directive.
		Updates(map[string]any{"result_directive": req.Directive})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update step group result directive")
	}

	return nil
}
