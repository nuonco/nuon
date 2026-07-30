package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type GetCurrentComponentConfigRequest struct {
	InstallID   string `validate:"required"`
	ComponentID string `validate:"required"`
}

// GetCurrentComponentConfig resolves the component's current config the same
// way the runner metadata handout does: the install's pinned app config first,
// falling back to the latest-configs view. A config version only carries ccc
// rows for components that CHANGED in that sync, so resolving by the pin alone
// silently loses every unchanged component — which is how a no-op sync turned
// the verified-deploy gate off.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 10s
// @by-field ComponentID
func (a *Activities) GetCurrentComponentConfig(ctx context.Context, req *GetCurrentComponentConfigRequest) (*app.ComponentConfigConnection, error) {
	var appConfigID string
	if err := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("app_config_id").
		Where("id = ?", req.InstallID).
		Scan(&appConfigID).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get install app config id")
	}

	if appConfigID != "" {
		var ccc app.ComponentConfigConnection
		err := a.db.WithContext(ctx).
			Where(app.ComponentConfigConnection{
				AppConfigID: appConfigID,
				ComponentID: req.ComponentID,
			}).
			First(&ccc).Error
		if err == nil {
			return &ccc, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err, "unable to get component config")
		}
	}

	var fallback app.ComponentConfigConnection
	err := a.db.WithContext(ctx).
		Scopes(
			scopes.WithDisableViews,
			scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
		).
		Where(app.ComponentConfigConnection{ComponentID: req.ComponentID}).
		First(&fallback).Error
	if err == nil {
		return &fallback, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return nil, errors.Wrap(err, "unable to get current component config")
}
