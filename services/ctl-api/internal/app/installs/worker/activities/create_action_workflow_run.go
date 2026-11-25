package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CreateActionWorkflowRunRequest struct {
	ActionWorkflowID        string `json:"action_workflow_id" validate:"required"`
	InstallActionWorkflowID string `json:"install_action_workflow_id" validate:"required"`
	InstallID               string `json:"install_id" validate:"required"`
	InstallWorkflowID       string `json:"install_workflow_id,omitempty"`
	AppConfigID             string `json:"app_config_id,omitempty"`
	ActionWorkflowConfigID  string `json:"action_workflow_config_id,omitempty"`

	TriggerType     app.ActionWorkflowTriggerType `json:"trigger_type" validate:"required"`
	TriggeredByID   string                        `json:"triggered_by_id"`
	TriggeredByType string                        `json:"triggered_by_type"`

	RunEnvVars map[string]*string `json:"run_env_vars"`
	Role       string             `json:"role,omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateActionWorkflowRun(ctx context.Context, req *CreateActionWorkflowRunRequest) (*app.InstallActionWorkflowRun, error) {
	run := func(a *Activities) (*app.InstallActionWorkflowRun, error) {
		return a.createActionWorkflowRun(ctx,
			req.ActionWorkflowID,
			req.InstallActionWorkflowID,
			req.InstallID,
			req.InstallWorkflowID,
			req.TriggerType,
			req.TriggeredByID,
			req.TriggeredByType,
			req.RunEnvVars,
			req.Role,
			req.ActionWorkflowConfigID,
		)
	}
	if req.AppConfigID == "" && req.ActionWorkflowConfigID == "" {
		return run(a)
	}
	if req.AppConfigID == "" || req.ActionWorkflowConfigID == "" {
		return nil, temporal.NewNonRetryableApplicationError("invalid prepared manual action target", "InvalidManualActionTarget", nil)
	}
	var result *app.InstallActionWorkflowRun
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var install app.Install
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "SHARE"}).
			Select("id", "app_config_id").Where(&app.Install{ID: req.InstallID}, "ID").First(&install).Error; err != nil {
			return err
		}
		if install.AppConfigID != req.AppConfigID {
			return temporal.NewNonRetryableApplicationError("install app config changed after the manual action was requested; start a new action run", "ManualActionConfigChanged", nil)
		}
		scoped := *a
		scoped.db = tx
		scoped.actionHelpers = a.actionHelpers.WithDB(tx)
		var err error
		result, err = run(&scoped)
		return err
	})
	return result, err
}

func (a *Activities) createActionWorkflowRun(ctx context.Context,
	actionWorkflowID string,
	installActionWorkflowID string,
	installID string,
	installWorkflowID string,
	triggerType app.ActionWorkflowTriggerType,
	triggeredByID string,
	triggeredByType string,
	runEnvVars map[string]*string,
	role string,
	actionWorkflowConfigID string,
) (*app.InstallActionWorkflowRun, error) {
	if triggerType == app.ActionWorkflowTriggerTypeAdHoc {
		var existingRun app.InstallActionWorkflowRun
		err := a.db.WithContext(ctx).
			Where("install_workflow_id = ? AND trigger_type = ?", installWorkflowID, app.ActionWorkflowTriggerTypeAdHoc).
			Preload("Steps").
			First(&existingRun).Error

		if err == nil {
			return &existingRun, nil
		}
	}

	install, err := a.getInstall(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	var cfg *app.ActionWorkflowConfig
	if actionWorkflowConfigID != "" {
		cfg, err = a.actionHelpers.GetActionWorkflowConfigByID(ctx, actionWorkflowConfigID)
	} else {
		cfg, err = a.actionHelpers.GetActionWorkflowConfig(ctx, actionWorkflowID, install.AppConfigID)
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflow config for install's current app config")
	}
	if actionWorkflowConfigID != "" && (cfg.AppConfigID != install.AppConfigID || cfg.ActionWorkflowID != actionWorkflowID || cfg.OrgID != install.OrgID) {
		return nil, temporal.NewNonRetryableApplicationError("prepared action config does not belong to this install's app config and action", "InvalidManualActionTarget", nil)
	}
	if actionWorkflowConfigID != "" {
		target, err := a.actionHelpers.GetInstallActionWorkflowForConfig(ctx, install, cfg)
		if err != nil {
			return nil, errors.Wrap(err, "unable to resolve prepared install action")
		}
		if target.ID != installActionWorkflowID {
			return nil, temporal.NewNonRetryableApplicationError("prepared install action does not match the selected action", "InvalidManualActionTarget", nil)
		}
	}

	steps := make([]app.InstallActionWorkflowRunStep, 0)
	for _, step := range cfg.Steps {
		steps = append(steps, app.InstallActionWorkflowRunStep{
			Status: app.InstallActionWorkflowRunStepStatusPending,
			StepID: generics.NewNullString(step.ID),
		})
	}

	newRun := app.InstallActionWorkflowRun{
		InstallActionWorkflowID: generics.NewNullString(installActionWorkflowID),
		InstallID:               installID,
		ActionWorkflowConfigID:  generics.NewNullString(cfg.ID),
		TriggerType:             triggerType,
		Status:                  app.InstallActionRunStatusQueued,
		StatusDescription:       "Queued",
		Steps:                   steps,
		RunEnvVars:              pgtype.Hstore(runEnvVars),
		TriggeredByID:           triggeredByID,
		TriggeredByType:         triggeredByType,
		Role:                    role,
		EnableKubeConfig:        cfg.EnableKubeConfig,
		KubernetesContextName:   cfg.KubernetesContextName,
	}

	if installWorkflowID != "" {
		newRun.InstallWorkflowID = generics.ToPtr(installWorkflowID)
	}

	res := a.db.WithContext(ctx).
		Create(&newRun)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create action workflow")
	}

	a.audit.Emit(ctx, audit.Event{
		Type:        audit.EventInstallActionWorkflowRun,
		Message:     "install action workflow run created",
		Outcome:     audit.OutcomeStarted,
		InstallID:   installID,
		SubjectID:   newRun.ID,
		SubjectType: "install_action_workflow_runs",
		Attrs: map[string]string{
			"action_workflow_run.id": newRun.ID,
			"action_workflow.id":     installActionWorkflowID,
			"action_workflow.config": cfg.ID,
			"trigger.type":           string(triggerType),
			"triggered_by.id":        triggeredByID,
			"triggered_by.type":      string(triggeredByType),
			"flow_workflow.id":       installWorkflowID,
		},
	})

	return &newRun, nil
}
