package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (h *Helpers) prepareRunbookTargets(ctx context.Context, tx *gorm.DB, selected *app.Install, association *app.InstallRunbook, cfg *app.RunbookConfig, selections []app.RunbookStepSelection) (*app.InstallExecutionTargets, error) {
	var install app.Install
	if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "SHARE"}).
		Where(&app.Install{ID: selected.ID, AppID: selected.AppID, OrgID: selected.OrgID}, "ID", "AppID", "OrgID").First(&install).Error; err != nil {
		return nil, err
	}
	if install.AppConfigID != selected.AppConfigID {
		return nil, stderr.ErrConflict{Err: fmt.Errorf("install app config changed"), Description: "install app config changed; start a new runbook run"}
	}
	if install.AppConfigID == "" {
		return nil, stderr.ErrUser{Err: fmt.Errorf("install has no app config"), Description: "select an app config before running a runbook"}
	}
	if association.InstallID != install.ID || association.OrgID != install.OrgID || association.Runbook.ID == "" || association.Runbook.AppID != install.AppID || association.Runbook.OrgID != install.OrgID || cfg.RunbookID != association.RunbookID || cfg.OrgID != install.OrgID || cfg.AppID != install.AppID || cfg.AppConfigID != install.AppConfigID {
		return nil, fmt.Errorf("runbook target does not belong to the install's app config and org")
	}
	var owner app.App
	if err := tx.Where(&app.App{ID: install.AppID, OrgID: install.OrgID}, "ID", "OrgID").First(&owner).Error; err != nil {
		return nil, err
	}
	appCfg, err := h.appsHelpers.WithDB(tx).GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		return nil, err
	}
	if appCfg.AppID != install.AppID || appCfg.OrgID != install.OrgID {
		return nil, fmt.Errorf("app config does not belong to the install's app and org")
	}
	targets := &app.InstallExecutionTargets{
		OrgID: install.OrgID, AppID: install.AppID, InstallID: install.ID, AppConfigID: appCfg.ID, RunbookConfigID: cfg.ID,
		Components: make(map[string]app.RunbookComponentTarget), Actions: make(map[string]app.RunbookActionTarget), StepComponents: make(map[string][]string),
		InstallComponents: make(map[string]string), TeardownComponents: make(map[string]app.RunbookTeardownTarget),
	}
	var installed []app.InstallComponent
	if err := tx.Preload("Component").Where(&app.InstallComponent{OrgID: install.OrgID, InstallID: install.ID}, "OrgID", "InstallID").Find(&installed).Error; err != nil {
		return nil, err
	}
	for _, component := range installed {
		if component.Component.OrgID == install.OrgID && component.Component.AppID == install.AppID {
			targets.InstallComponents[component.ComponentID] = component.ID
		}
	}
	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.OrgID != install.OrgID || ccc.Component.ID != ccc.ComponentID || ccc.Component.OrgID != install.OrgID || ccc.Component.AppID != install.AppID {
			return nil, fmt.Errorf("component config does not belong to the install's app and org")
		}
		targets.Components[ccc.ComponentID] = app.RunbookComponentTarget{ConfigID: ccc.ID, Name: ccc.Component.Name}
	}
	var actions []app.InstallActionWorkflow
	if err := tx.Preload("ActionWorkflow").Where(&app.InstallActionWorkflow{OrgID: install.OrgID, InstallID: install.ID}, "OrgID", "InstallID").Find(&actions).Error; err != nil {
		return nil, err
	}
	for _, action := range actions {
		for _, actionCfg := range appCfg.ActionWorkflowConfigs {
			if actionCfg.ActionWorkflowID != action.ActionWorkflowID {
				continue
			}
			if action.ActionWorkflow.ID == "" || action.ActionWorkflow.OrgID != install.OrgID || action.ActionWorkflow.AppID != install.AppID || actionCfg.OrgID != install.OrgID || actionCfg.AppID != install.AppID || actionCfg.AppConfigID != appCfg.ID {
				return nil, fmt.Errorf("action does not belong to the install's app config and org")
			}
			targets.Actions[action.ActionWorkflowID] = app.RunbookActionTarget{InstallActionWorkflowID: action.ID, ConfigID: actionCfg.ID}
		}
	}
	disabled := make(map[string]bool)
	for _, selection := range selections {
		if !selection.Enabled {
			disabled[selection.StepID] = true
		}
	}
	for _, step := range cfg.Steps {
		if disabled[step.ID] {
			continue
		}
		if step.OrgID != install.OrgID || step.RunbookConfigID != cfg.ID {
			return nil, fmt.Errorf("step does not belong to runbook config")
		}
		switch step.Type {
		case app.RunbookStepTypeComponentDeploy, app.RunbookStepTypeComponentTearDown:
			var component app.Component
			if err := tx.Where(&app.Component{OrgID: install.OrgID, AppID: install.AppID, Name: step.ComponentName}, "OrgID", "AppID", "Name").First(&component).Error; err != nil {
				return nil, fmt.Errorf("resolve runbook component %s: %w", step.ComponentName, err)
			}
			ids := []string{component.ID}
			if step.Type == app.RunbookStepTypeComponentDeploy && step.DeployDependents {
				ids, err = h.appsHelpers.WithDB(tx).GetConfigComponentDeployOrder(ctx, appCfg, component.ID)
			} else if step.Type == app.RunbookStepTypeComponentTearDown && step.TearDownDependents {
				ids, err = h.appsHelpers.WithDB(tx).GetReverseConfigComponentDeployOrder(ctx, appCfg, component.ID)
			}
			if err != nil {
				return nil, err
			}
			for _, id := range ids {
				if targets.InstallComponents[id] == "" {
					return nil, fmt.Errorf("runbook component %s is not installed in this org", id)
				}
				if step.Type == app.RunbookStepTypeComponentTearDown {
					target, err := prepareRunbookTeardownTarget(tx, &install, targets.InstallComponents[id])
					if err != nil {
						return nil, err
					}
					targets.TeardownComponents[id] = *target
				} else if _, ok := targets.Components[id]; !ok {
					return nil, fmt.Errorf("runbook component %s is not in the install's app config", id)
				}
			}
			targets.StepComponents[step.ID] = ids
		case app.RunbookStepTypeAction:
			if actionID := step.ActionWorkflowID.ValueString(); actionID != "" {
				if _, ok := targets.Actions[actionID]; !ok {
					return nil, fmt.Errorf("runbook action %s is not installed in this app config", actionID)
				}
			}
		case app.RunbookStepTypeSandboxReprovision, app.RunbookStepTypeSandboxDeprovision:
			var sandbox app.InstallSandbox
			if err := tx.Where(&app.InstallSandbox{InstallID: install.ID, OrgID: install.OrgID}, "InstallID", "OrgID").First(&sandbox).Error; err != nil {
				return nil, err
			}
			targets.InstallSandboxID = sandbox.ID
			if appCfg.SandboxConfig.ID == "" || appCfg.SandboxConfig.OrgID != install.OrgID || appCfg.SandboxConfig.AppID != install.AppID {
				return nil, fmt.Errorf("sandbox config does not belong to the install's app and org")
			}
			if step.Type == app.RunbookStepTypeSandboxReprovision && !step.SkipComponentDeploys && !step.PlanOnly {
				targets.ComponentOrder, err = h.appsHelpers.WithDB(tx).GetConfigDefaultComponentOrder(ctx, appCfg)
				if err != nil {
					return nil, err
				}
				for _, id := range targets.ComponentOrder {
					if targets.InstallComponents[id] == "" {
						return nil, fmt.Errorf("runbook component %s is not installed in this org", id)
					}
				}
			}
		}
	}
	return targets, nil
}

func prepareRunbookTeardownTarget(tx *gorm.DB, install *app.Install, installComponentID string) (*app.RunbookTeardownTarget, error) {
	var component app.InstallComponent
	if err := tx.Preload("Component").Where(&app.InstallComponent{ID: installComponentID, InstallID: install.ID, OrgID: install.OrgID}, "ID", "InstallID", "OrgID").First(&component).Error; err != nil {
		return nil, err
	}
	target := &app.RunbookTeardownTarget{Name: component.Component.Name, Type: component.Component.Type}
	var deploy app.InstallDeploy
	query := tx.Where(&app.InstallDeploy{InstallComponentID: component.ID, OrgID: install.OrgID, Type: app.InstallDeployTypeApply}, "InstallComponentID", "OrgID", "Type").Order("created_at DESC").Session(&gorm.Session{})
	err := query.Where("status IN ?", app.AppliedDeployStatuses()).First(&deploy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = query.First(&deploy).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return target, nil
	}
	if err != nil {
		return nil, err
	}
	var build app.ComponentBuild
	if err := tx.Where(&app.ComponentBuild{ID: deploy.ComponentBuildID, OrgID: install.OrgID}, "ID", "OrgID").First(&build).Error; err != nil {
		return nil, err
	}
	var cfg app.ComponentConfigConnection
	if err := tx.Preload("TerraformModuleComponentConfig").Preload("HelmComponentConfig").Preload("ExternalImageComponentConfig").Preload("JobComponentConfig").Preload("KubernetesManifestComponentConfig").Preload("PulumiComponentConfig").
		Where(&app.ComponentConfigConnection{ID: build.ComponentConfigConnectionID, ComponentID: component.ComponentID, OrgID: install.OrgID}, "ID", "ComponentID", "OrgID").First(&cfg).Error; err != nil {
		return nil, err
	}
	target.Type = cfg.Type
	target.ConfigID = cfg.ID
	return target, nil
}
