package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type SetInstallDeployPlanCompositeErrorRequest struct {
	InstallDeployID string `validate:"required"`
	ComponentName   string
	Stage           string
	Detail          string
}

// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) SetInstallDeployPlanCompositeError(ctx context.Context, req SetInstallDeployPlanCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("install_deploy_id", req.InstallDeployID))

	var data *compositeerrors.CompositeErrorData
	if req.Detail != "" {
		componentName := req.ComponentName
		if componentName == "" {
			var err error
			componentName, err = a.componentNameForDeploy(ctx, req.InstallDeployID)
			if err != nil {
				l.Warn("unable to resolve component name for deploy", zap.Error(err))
			}
		}

		appSyncDisabled, err := a.deployInstallAppSyncDisabled(ctx, req.InstallDeployID)
		if err != nil {
			l.Warn("unable to check disable-app-sync for deploy", zap.Error(err))
		}

		data, err = compositeerrors.New(
			&deployerrors.DeployPlanRenderError{
				ComponentName:   componentName,
				Stage:           req.Stage,
				Detail:          req.Detail,
				AppSyncDisabled: appSyncDisabled,
			},
			compositeerrors.WithSource("install_deploys", req.InstallDeployID),
		)
		if err != nil {
			return fmt.Errorf("unable to build deploy plan render composite error: %w", err)
		}
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{ID: req.InstallDeployID}).
		Select("composite_error").
		Updates(app.InstallDeploy{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to set install deploy plan composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no install deploy found for id %s: %w", req.InstallDeployID, gorm.ErrRecordNotFound)
	}

	l.Info("updated install deploy plan composite error")
	return nil
}

// deployInstallAppSyncDisabled reports whether disable-app-sync is enabled
// for the org that owns the install behind the given install deploy.
func (a *Activities) deployInstallAppSyncDisabled(ctx context.Context, installDeployID string) (bool, error) {
	var installID string
	if err := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{}).
		Select("install_components.install_id").
		Joins("join install_components on install_components.id = install_deploys.install_component_id").
		Where(app.InstallDeploy{ID: installDeployID}).
		Scan(&installID).Error; err != nil {
		return false, fmt.Errorf("unable to get install for install deploy %s: %w", installDeployID, err)
	}

	var install app.Install
	if err := a.db.WithContext(ctx).
		Select("org_id").
		Where(app.Install{ID: installID}).
		First(&install).Error; err != nil {
		return false, fmt.Errorf("unable to get install %s: %w", installID, err)
	}
	return a.features.OrgHasFeature(ctx, install.OrgID, app.OrgFeatureDisableAppSync)
}

func (a *Activities) componentNameForDeploy(ctx context.Context, installDeployID string) (string, error) {
	var name string
	err := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{}).
		Select("components.name").
		Joins("join install_components on install_components.id = install_deploys.install_component_id").
		Joins("join components on components.id = install_components.component_id").
		Where(app.InstallDeploy{ID: installDeployID}).
		Scan(&name).Error
	if err != nil {
		return "", fmt.Errorf("unable to get component name for install deploy %s: %w", installDeployID, err)
	}
	return name, nil
}
