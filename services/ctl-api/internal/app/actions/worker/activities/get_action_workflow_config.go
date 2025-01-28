package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetActionWorkflowConfigRequest struct {
	ConfigID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ConfigID
func (a *Activities) GetActionWorkflowConfig(ctx context.Context, req *GetActionWorkflowConfigRequest) (*app.ActionWorkflowConfig, error) {
	return a.getActionWorkflowConfig(ctx, req.ConfigID)
}

func (a *Activities) getActionWorkflowConfig(ctx context.Context, configID string) (*app.ActionWorkflowConfig, error) {
	cfg := app.ActionWorkflowConfig{}
	res := a.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Triggers").
		Preload("Steps").
		First(&cfg, "id = ?", configID)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get action workflow config")
	}

	return &cfg, nil
}
