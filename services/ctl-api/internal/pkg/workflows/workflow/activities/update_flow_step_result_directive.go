package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateFlowStepResultDirectiveRequest struct {
	StepID    string `validate:"required"`
	Directive string // empty string is valid (used to clear the directive)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepResultDirective(ctx context.Context, req UpdateFlowStepResultDirectiveRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{ID: req.StepID}).
		Updates(map[string]any{"result_directive": req.Directive})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update step result directive")
	}

	return nil
}
