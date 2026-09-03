package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func (a *Activities) checkComponentBuilds(ctx context.Context, req InstallPreflightRequest) ([]*compositeerrors.CompositeErrorData, error) {
	findings := make([]*compositeerrors.CompositeErrorData, 0)
	for _, planned := range req.PlannedComponentBuilds {
		finding, err := a.checkComponentBuild(ctx, req, planned)
		if err != nil {
			return nil, err
		}
		if finding != nil {
			findings = append(findings, finding)
		}
	}
	return findings, nil
}

func (a *Activities) checkComponentBuild(ctx context.Context, req InstallPreflightRequest, planned PlannedComponentBuild) (*compositeerrors.CompositeErrorData, error) {
	build, componentID, componentName, err := a.resolvePlannedComponentBuild(ctx, req, planned)
	if err != nil {
		return nil, err
	}

	if build != nil {
		switch build.Status {
		case app.ComponentBuildStatusActive:
			return nil, nil
		case app.ComponentBuildStatusQueued,
			app.ComponentBuildStatusPlanning,
			app.ComponentBuildStatusBuilding:
			if planned.WaitForBuild {
				return nil, nil
			}
		}
		if componentID == "" {
			componentID = build.ComponentID
		}
		if componentName == "" {
			componentName = build.ComponentName
		}
	}

	if componentName == "" && componentID != "" {
		var component app.Component
		if err := a.db.WithContext(ctx).
			Select("id", "name").
			Where(app.Component{ID: componentID}).
			First(&component).Error; err != nil {
			return nil, fmt.Errorf("unable to get planned component: %w", err)
		}
		componentName = component.Name
	}

	reason := deployerrors.ComponentBuildUnavailableReasonMissing
	if build != nil && (build.Status == app.ComponentBuildStatusError || build.Status == app.ComponentBuildStatusPolicyFailed) {
		reason = deployerrors.ComponentBuildUnavailableReasonFailed
	}
	errorData := &deployerrors.ComponentBuildUnavailableError{
		Reason:        reason,
		ComponentID:   componentID,
		ComponentName: componentName,
	}
	sourceType := (&app.Workflow{}).TableName()
	sourceID := req.FlowID
	if build != nil {
		errorData.BuildID = build.ID
		errorData.BuildStatus = string(build.Status)
		errorData.BuildStatusDescription = build.StatusDescription
		sourceType = "component_builds"
		sourceID = build.ID
	}

	finding, err := compositeerrors.New(errorData, compositeerrors.WithSource(sourceType, sourceID))
	if err != nil {
		return nil, fmt.Errorf("unable to build component preflight error: %w", err)
	}
	return finding, nil
}

func (a *Activities) resolvePlannedComponentBuild(ctx context.Context, req InstallPreflightRequest, planned PlannedComponentBuild) (*app.ComponentBuild, string, string, error) {
	componentID := planned.ComponentID
	componentName := ""

	switch {
	case planned.DeployID != "":
		var deploy app.InstallDeploy
		if err := a.db.WithContext(ctx).
			Preload("ComponentBuild").
			Preload("InstallComponent.Component").
			Where(app.InstallDeploy{ID: planned.DeployID}).
			First(&deploy).Error; err != nil {
			return nil, "", "", fmt.Errorf("unable to get planned component deploy: %w", err)
		}
		if componentID == "" {
			componentID = deploy.ComponentID
		}
		componentName = deploy.ComponentName
		if componentName == "" {
			componentName = deploy.InstallComponent.Component.Name
		}
		return &deploy.ComponentBuild, componentID, componentName, nil
	case planned.BuildID != "":
		var build app.ComponentBuild
		err := a.db.WithContext(ctx).
			Select("id", "status", "status_description", "component_config_connection_id").
			Where(app.ComponentBuild{ID: planned.BuildID}).
			First(&build).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, componentID, componentName, nil
		}
		if err != nil {
			return nil, "", "", fmt.Errorf("unable to get planned component build: %w", err)
		}
		return &build, componentID, componentName, nil
	}

	connectionID := planned.ComponentConfigConnectionID
	if connectionID == "" {
		desiredAppConfigID := req.DesiredAppConfigID
		if desiredAppConfigID == "" {
			var install app.Install
			if err := a.db.WithContext(ctx).
				Select("id", "app_config_id").
				Where(app.Install{ID: req.InstallID}).
				First(&install).Error; err != nil {
				return nil, "", "", fmt.Errorf("unable to get install: %w", err)
			}
			desiredAppConfigID = install.AppConfigID
		}

		appConfig, err := a.appsHelpers.GetFullAppConfig(ctx, desiredAppConfigID, false)
		if err != nil {
			return nil, "", "", fmt.Errorf("unable to get desired app config: %w", err)
		}
		connectionID = componentConfigConnectionID(appConfig, componentID)
		if connectionID == "" {
			return nil, componentID, componentName, nil
		}
	}

	build, err := a.GetComponentBuildForConfigConnection(ctx, GetComponentBuildForConfigConnectionRequest{
		ComponentConfigConnectionID: connectionID,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("unable to get component build for planned configuration: %w", err)
	}
	return build, componentID, componentName, nil
}

func componentConfigConnectionID(appConfig *app.AppConfig, componentID string) string {
	for _, connection := range appConfig.ComponentConfigConnections {
		if connection.ComponentID == componentID {
			return connection.ID
		}
	}
	return ""
}
