package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetActionWorkflowConfig struct {
	ActionWorkflowID string `validate:"required"`
	AppConfigID      string `validate:"required"`
}

// @temporal-gen activity
// @by-id ActionWorkflowID
func (a *Activities) GetActionWorkflowConfig(ctx context.Context, req *GetActionWorkflowConfig) (*app.ActionWorkflowConfig, error) {
	return a.actionHelpers.GetActionWorkflowConfig(ctx, req.ActionWorkflowID, req.AppConfigID)
}
