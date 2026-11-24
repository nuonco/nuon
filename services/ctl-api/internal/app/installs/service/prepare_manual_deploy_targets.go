package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) prepareManualDeployTargets(ctx context.Context, db *gorm.DB, target *preparedInstallDeploy, req *CreateInstallDeployRequest) (*app.InstallExecutionTargets, error) {
	install := &target.Install
	cfg, err := s.appsHelpers.WithDB(db).GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		return nil, err
	}
	if cfg.AppID != install.AppID || cfg.OrgID != install.OrgID {
		return nil, fmt.Errorf("app config does not belong to the install's app and org")
	}
	targets := &app.InstallExecutionTargets{
		InstallID: install.ID, AppID: install.AppID, OrgID: install.OrgID, AppConfigID: cfg.ID,
		PrimaryComponentID: target.InstallComponent.ComponentID, PrimaryBuildID: target.Build.ID,
		PrimaryType: target.Build.ComponentConfigConnection.Type,
		Components:  make(map[string]app.RunbookComponentTarget), InstallComponents: make(map[string]string), Actions: make(map[string]app.RunbookActionTarget),
	}
	for _, ccc := range cfg.ComponentConfigConnections {
		if ccc.OrgID != install.OrgID || ccc.Component.ID != ccc.ComponentID || ccc.Component.OrgID != install.OrgID || ccc.Component.AppID != install.AppID {
			return nil, fmt.Errorf("component config does not belong to the install's app and org")
		}
		targets.Components[ccc.ComponentID] = app.RunbookComponentTarget{ConfigID: ccc.ID, Name: ccc.Component.Name}
	}
	var components []app.InstallComponent
	if err := db.WithContext(ctx).Preload("Component").Where(&app.InstallComponent{InstallID: install.ID, OrgID: install.OrgID}, "InstallID", "OrgID").Find(&components).Error; err != nil {
		return nil, err
	}
	for _, component := range components {
		if component.Component.AppID == install.AppID && component.Component.OrgID == install.OrgID {
			targets.InstallComponents[component.ComponentID] = component.ID
		}
	}
	if err := targets.CheckComponent(target.InstallComponent.ComponentID, &target.InstallComponent); err != nil {
		return nil, err
	}
	if req.DeployDependents {
		order, err := s.appsHelpers.WithDB(db).GetConfigComponentDeployOrder(ctx, cfg, targets.PrimaryComponentID)
		if err != nil {
			return nil, err
		}
		targets.ComponentOrder = generics.SliceAfterValue(order, targets.PrimaryComponentID)
		for _, id := range targets.ComponentOrder {
			if targets.InstallComponents[id] == "" || targets.Components[id].ConfigID == "" {
				return nil, fmt.Errorf("dependent component %s is not installed in this app config and org", id)
			}
		}
	}
	if !req.PlanOnly {
		var actions []app.InstallActionWorkflow
		if err := db.WithContext(ctx).Preload("ActionWorkflow").Where(&app.InstallActionWorkflow{InstallID: install.ID, OrgID: install.OrgID}, "InstallID", "OrgID").Find(&actions).Error; err != nil {
			return nil, err
		}
		for _, action := range actions {
			for _, actionCfg := range cfg.ActionWorkflowConfigs {
				if action.ActionWorkflowID != actionCfg.ActionWorkflowID {
					continue
				}
				selected := false
				for _, id := range append([]string{targets.PrimaryComponentID}, targets.ComponentOrder...) {
					if actionCfg.HasComponentTrigger(app.ActionWorkflowTriggerTypePreDeployComponent, id) || actionCfg.HasComponentTrigger(app.ActionWorkflowTriggerTypePostDeployComponent, id) {
						selected = true
						break
					}
				}
				if !selected {
					continue
				}
				if action.ActionWorkflow.ID == "" || action.ActionWorkflow.AppID != install.AppID || action.ActionWorkflow.OrgID != install.OrgID || actionCfg.AppID != install.AppID || actionCfg.OrgID != install.OrgID || actionCfg.AppConfigID != cfg.ID {
					return nil, fmt.Errorf("action does not belong to the install's app config and org")
				}
				targets.Actions[action.ActionWorkflowID] = app.RunbookActionTarget{InstallActionWorkflowID: action.ID, ConfigID: actionCfg.ID}
			}
		}
	}
	return targets, nil
}

func manualDeployMetadata(target *preparedInstallDeploy, deploy *app.InstallDeploy, req *CreateInstallDeployRequest, name string) (map[string]string, error) {
	blob, err := json.Marshal(target.Targets)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		app.WorkflowMetadataKeyWorkflowNameSuffix: name,
		"install_deploy_id":                       deploy.ID,
		"deploy_dependents":                       strconv.FormatBool(req.DeployDependents),
		"deploy_dependencies":                     strconv.FormatBool(req.DeployDependencies),
		app.ManualDeployTargetsMetadataKey:        string(blob),
	}, nil
}
