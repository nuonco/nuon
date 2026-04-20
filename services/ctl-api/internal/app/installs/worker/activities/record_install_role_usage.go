package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

type RecordInstallRoleUsageRequest struct {
	InstallID     string                        `validate:"required" temporaljson:"install_id"`
	RunnerJobID   string                        `validate:"required" temporaljson:"runner_job_id"`
	RoleSelection *operationroles.RoleSelection `temporaljson:"role_selection"`
}

// @temporal-gen-v2 activity
func (a *Activities) RecordInstallRoleUsage(ctx context.Context, req *RecordInstallRoleUsageRequest) error {
	if req.RoleSelection == nil {
		return nil
	}
	if req.RoleSelection.UnrenderedRoleName == "" && req.RoleSelection.RoleName == "" {
		return nil
	}

	var install app.Install
	res := a.db.WithContext(ctx).
		Where("id = ?", req.InstallID).
		First(&install)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}

	var installRoles []app.InstallRoles
	res = a.db.WithContext(ctx).
		Preload("AppRoleConfig", func(db *gorm.DB) *gorm.DB {
			return db.Where("app_config_id = ?", install.AppConfigID)
		}).
		Where("install_id = ?", req.InstallID).
		Find(&installRoles)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}

	installState, err := a.helpers.GetInstallState(ctx, req.InstallID, false, false)
	var stateMap map[string]any
	if err == nil && installState != nil {
		stateMap, _ = installState.AsMap()
	}

	var matchedRoleID string
	for _, ir := range installRoles {
		name := ir.AppRoleConfig.Name
		if name == "" {
			continue
		}
		if name == req.RoleSelection.UnrenderedRoleName {
			matchedRoleID = ir.ID
			break
		}
		if stateMap != nil {
			if rendered, renderErr := render.RenderV2(name, stateMap); renderErr == nil {
				if rendered == req.RoleSelection.RoleName || rendered == req.RoleSelection.UnrenderedRoleName {
					matchedRoleID = ir.ID
					break
				}
			}
		}
	}
	if matchedRoleID == "" {
		return nil
	}

	usage := app.InstallRoleUsage{
		InstallRoleID:      matchedRoleID,
		RunnerJobID:        req.RunnerJobID,
		RoleName:           req.RoleSelection.RoleName,
		RoleSource:         string(req.RoleSelection.Source),
		RoleSelectionTrace: app.RunnerJobPermissionTrace(req.RoleSelection.Trace),
	}

	if err := a.db.WithContext(ctx).Create(&usage).Error; err != nil {
		return generics.TemporalGormError(err)
	}

	return nil
}
