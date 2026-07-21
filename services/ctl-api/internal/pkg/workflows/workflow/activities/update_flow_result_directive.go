package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateFlowResultDirectiveRequest struct {
	FlowID    string `validate:"required"`
	Directive string // empty string is valid (used to clear the directive)
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowResultDirective(ctx context.Context, req UpdateFlowResultDirectiveRequest) error {
	flow := app.Workflow{ID: req.FlowID}
	res := a.db.WithContext(ctx).
		Model(&flow).
		Clauses(clause.Returning{}).
		Where(app.Workflow{ID: req.FlowID}).
		// Must use map, not struct — GORM's struct-based Updates() skips zero-value
		// fields, so Updates(app.Workflow{ResultDirective: ""}) would be a no-op
		// when clearing the directive.
		Updates(map[string]any{"result_directive": req.Directive})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update workflow result directive")
	}

	fields := []zap.Field{
		zap.String("workflow_id", req.FlowID),
		zap.String("workflow_type", string(flow.Type)),
		zap.String("org_id", flow.OrgID),
	}
	if flow.OwnerType == "installs" {
		fields = append(fields, zap.String("install_id", flow.OwnerID))
	}
	a.logDirective(ctx, "workflow.directive", req.Directive, fields...)

	return nil
}
