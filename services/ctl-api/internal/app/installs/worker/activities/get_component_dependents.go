package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentDependents struct {
	AppID           string `json:"app_id" validate:"required"`
	ComponentRootID string `json:"component_root_id" validate:"required"`
	ConfigVersion   int    `json:"config_version" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetComponentDependents(ctx context.Context, req GetComponentDependents) ([]app.Component, error) {
	comps, err := a.appsHelpers.GetInvertedDependentByComponentConfigVersion(ctx, req.AppID, req.ConfigVersion, req.ComponentRootID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dependent components")
	}

	return comps, nil
}
