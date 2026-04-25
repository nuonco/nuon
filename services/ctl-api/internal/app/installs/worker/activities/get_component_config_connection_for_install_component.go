package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetComponentConfigConnectionForInstallComponentRequest struct {
	InstallComponentID string `validate:"required"`
	ComponentID        string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetComponentConfigConnectionForInstallComponent(ctx context.Context, req GetComponentConfigConnectionForInstallComponentRequest) (*app.ComponentConfigConnection, error) {
	var ccc app.ComponentConfigConnection
	viewOrTable := views.TableOrViewName(a.db, &app.ComponentConfigConnection{}, "")
	res := a.db.WithContext(ctx).
		Table(viewOrTable).
		Joins("JOIN installs ON installs.app_config_id = "+viewOrTable+".app_config_id").
		Joins("JOIN install_components ON install_components.install_id = installs.id").
		Where("install_components.id = ?", req.InstallComponentID).
		Where(viewOrTable+".component_id = ?", req.ComponentID).
		First(&ccc)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component config connection for install component: %w", res.Error)
	}

	return &ccc, nil
}
