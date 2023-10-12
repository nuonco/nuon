package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminUpdateSandboxRequest struct{}

//	@BasePath	/v1/installs
//
// Update an install to the latest sandbox release
//
//		@Summary	update an install to the latest sandbox
//		@Schemes
//		@Description	update and app to the latest sandbox
//	  @Param			install_id	path	string					true	"app ID"
//		@Param			req			body	AdminUpdateSandboxRequest	true	"Input"
//		@Tags			installs/admin
//		@Accept			json
//		@Produce		json
//		@Success		200	{boolean}	true
//		@Router			/v1/installs/{install_id}/admin-update-sandbox [POST]
func (s *service) AdminUpdateSandbox(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AdminUpdateSandboxRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	app, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	sandboxRelease, err := s.getLatestSandbox(ctx, app.SandboxRelease.SandboxID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest sandbox release: %w", err))
		return
	}

	if _, err := s.updateInstallSandbox(ctx, installID, sandboxRelease.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install sandbox: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getLatestSandbox(ctx context.Context, sandboxID string) (*app.SandboxRelease, error) {
	release := app.SandboxRelease{}

	res := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(1).
		First(&release, "sandbox_id = ?", sandboxID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox releases: %w", res.Error)
	}

	return &release, nil
}

func (s *service) updateInstallSandbox(ctx context.Context, installID string, sandboxReleaseID string) (*app.Install, error) {
	currentInstall := app.Install{
		ID: installID,
	}

	res := s.db.WithContext(ctx).
		Preload("SandboxRelease").
		Model(&currentInstall).
		Updates(app.App{SandboxReleaseID: sandboxReleaseID})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update install: %w", res.Error)
	}

	return &currentInstall, nil
}
