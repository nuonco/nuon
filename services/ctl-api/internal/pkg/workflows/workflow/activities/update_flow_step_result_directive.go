package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateFlowStepResultDirectiveRequest struct {
	StepID    string `validate:"required"`
	Directive string // empty string is valid (used to clear the directive)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepResultDirective(ctx context.Context, req UpdateFlowStepResultDirectiveRequest) error {
	step := app.WorkflowStep{ID: req.StepID}
	res := a.db.WithContext(ctx).
		Model(&step).
		Clauses(clause.Returning{}).
		Where(app.WorkflowStep{ID: req.StepID}).
		Updates(map[string]any{"result_directive": req.Directive})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update step result directive")
	}

	a.logDirective(ctx, "step.directive", req.Directive,
		zap.String("step_id", req.StepID),
		zap.String("workflow_id", step.InstallWorkflowID),
		zap.String("org_id", step.OrgID),
		zap.String("step_name", step.Name),
		zap.Int("step_idx", step.Idx),
		zap.Int("group_idx", step.GroupIdx),
	)

	return nil
}
