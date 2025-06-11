package helpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (h *Helpers) GetActionWorkflowConfig(ctx context.Context, actionWorkflowID, appConfigId string) (*app.ActionWorkflowConfig, error) {
	if actionWorkflowID == "" || appConfigId == "" {
		return nil, errors.New("action workflow ID and app config ID are required to be non empty")
	}

	var actionWorkflowConfig app.ActionWorkflowConfig

	res := h.db.WithContext(ctx).
		Where(app.ActionWorkflowConfig{
			ActionWorkflowID: actionWorkflowID,
			AppConfigID:      appConfigId,
		}).
		Preload("Triggers").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		Order("created_at desc").
		First(&actionWorkflowConfig)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get action workflow config for given app config ID")
	}

	return &actionWorkflowConfig, nil
}

func (h *Helpers) GetActionWorkflowConfigByID(ctx context.Context, actionWorkflowConfigID string) (*app.ActionWorkflowConfig, error) {
	if actionWorkflowConfigID == "" {
		return nil, errors.New("action workflow config ID is required to be non empty")
	}

	actionWorkflowConfig := app.ActionWorkflowConfig{
		ID: actionWorkflowConfigID,
	}

	res := h.db.WithContext(ctx).
		Preload("Triggers").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		First(&actionWorkflowConfig)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get action workflow config")
	}

	return &actionWorkflowConfig, nil
}
