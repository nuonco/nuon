package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetActionWorkflowLatestConfig struct {
	ActionWorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ActionWorkflowID
func (a *Activities) GetActionWorkflowLatestConfig(ctx context.Context, req *GetActionWorkflowLatestConfig) (*app.ActionWorkflowConfig, error) {
	return a.getActionWorkflowLatestConfig(ctx, req.ActionWorkflowID)
}

func (a *Activities) getActionWorkflowLatestConfig(ctx context.Context, actionWorkflowID string) (*app.ActionWorkflowConfig, error) {
	var actionWorkflowConfig app.ActionWorkflowConfig

	res := a.db.WithContext(ctx).
		Where(app.ActionWorkflowConfig{
			ActionWorkflowID: actionWorkflowID,
		}).
		Preload("Triggers").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		Order("created_at desc").
		First(&actionWorkflowConfig)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get action workflow latest config")
	}

	return &actionWorkflowConfig, nil
}
