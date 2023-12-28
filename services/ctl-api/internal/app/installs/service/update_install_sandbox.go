package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type AdminUpdateSandboxRequest struct{}

// @ID AdminUpdateInstallSandbox
// @Summary	update an install to the latest sandbox
// @Description	update_install_sandbox.md
// @Param			install_id	path	string						true	"app ID"
// @Param			req			body	AdminUpdateSandboxRequest	true	"Input"
// @Tags			installs/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/installs/{install_id}/admin-update-sandbox [POST]
func (s *service) AdminUpdateSandbox(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AdminUpdateSandboxRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	appSandboxConfig, err := s.getLatestAppSandboxConfig(ctx, install.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest app sandbox config: %w", err))
		return
	}

	if _, err := s.updateInstallSandbox(ctx, installID, appSandboxConfig.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install sandbox: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getLatestAppSandboxConfig(ctx context.Context, appID string) (*app.AppSandboxConfig, error) {
	parentApp := app.App{}

	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app sandbox configs: %w", res.Error)
	}
	if len(parentApp.AppSandboxConfigs) < 1 {
		return nil, fmt.Errorf("app does not have any sandbox configs: %w", gorm.ErrRecordNotFound)
	}

	return &parentApp.AppSandboxConfigs[0], nil
}

func (s *service) updateInstallSandbox(ctx context.Context, installID string, sandboxReleaseID string) (*app.Install, error) {
	currentInstall := app.Install{
		ID: installID,
	}

	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Model(&currentInstall).
		Updates(app.App{})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update install: %w", res.Error)
	}

	return &currentInstall, nil
}
