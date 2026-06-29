package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallRolesForInstallRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

type GetInstallRolesForInstallResponse struct {
	Roles []app.InstallRoles `json:"roles"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallRolesForInstall(ctx context.Context, req GetInstallRolesForInstallRequest) (*GetInstallRolesForInstallResponse, error) {
	var roles []app.InstallRoles
	if res := a.db.WithContext(ctx).
		Preload("AppRoleConfig").
		Where(app.InstallRoles{InstallID: req.InstallID}).
		Find(&roles); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install roles")
	}

	return &GetInstallRolesForInstallResponse{Roles: roles}, nil
}
