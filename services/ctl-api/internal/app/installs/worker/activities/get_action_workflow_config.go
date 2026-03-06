package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetActionWorkflowConfig struct {
	ActionWorkflowID string `validate:"required"`
	AppConfigID      string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ActionWorkflowID
func (a *Activities) GetActionWorkflowConfig(ctx context.Context, req *GetActionWorkflowConfig) (*app.ActionWorkflowConfig, error) {
	return a.actionHelpers.GetActionWorkflowConfig(ctx, req.ActionWorkflowID, req.AppConfigID)
}
