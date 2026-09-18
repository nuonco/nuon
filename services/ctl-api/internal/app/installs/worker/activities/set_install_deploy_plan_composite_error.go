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

		var err error
		data, err = compositeerrors.New(
			&deployerrors.DeployPlanRenderError{
				ComponentName: componentName,
				Stage:         req.Stage,
				Detail:        req.Detail,
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
