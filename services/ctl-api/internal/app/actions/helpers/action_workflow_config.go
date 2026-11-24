package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

func (h *Helpers) GetInstallActionWorkflowForConfig(ctx context.Context, install *app.Install, cfg *app.ActionWorkflowConfig) (*app.InstallActionWorkflow, error) {
	if cfg.OrgID != install.OrgID || cfg.AppConfigID != install.AppConfigID {
		return nil, errors.New("action config does not belong to the install's app config and org")
	}
	var appConfig app.AppConfig
	if err := h.db.WithContext(ctx).
		Where(&app.AppConfig{ID: cfg.AppConfigID, AppID: install.AppID, OrgID: install.OrgID}, "ID", "AppID", "OrgID").
		First(&appConfig).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get install app config")
	}
	var target app.InstallActionWorkflow
	if err := h.db.WithContext(ctx).Preload("ActionWorkflow.App").
		Where(&app.InstallActionWorkflow{InstallID: install.ID, ActionWorkflowID: cfg.ActionWorkflowID, OrgID: install.OrgID}, "InstallID", "ActionWorkflowID", "OrgID").
		First(&target).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get install action")
	}
	action := &target.ActionWorkflow
	if action.ID == "" || action.OrgID != install.OrgID || action.AppID != install.AppID || action.App.ID == "" || action.App.OrgID != install.OrgID {
		return nil, errors.New("action does not belong to the install's app and org")
	}
	return &target, nil
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
