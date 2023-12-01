package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminUpdateSandboxRequest struct{}

//	@BasePath	/v1/apps
//
// Update an app to the latest sandbox release
//
//	@Summary	update an app to the latest sandbox
//	@Schemes
//	@Description	update and app to the latest sandbox
//	@Param			app_id	path	string						true	"app ID"
//	@Param			req		body	AdminUpdateSandboxRequest	true	"Input"
//	@Tags			apps/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/apps/{app_id}/admin-update-sandbox [POST]
func (s *service) AdminUpdateSandbox(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req AdminUpdateSandboxRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	app, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	if len(app.AppSandboxConfigs) < 1 {
		ctx.Error(fmt.Errorf("no app sandbox configs found"))
		return
	}

	sandboxName := app.AppSandboxConfigs[0].SandboxRelease.Sandbox.Name
	if sandboxName == "" {
		ctx.Error(fmt.Errorf("no built in app sandbox found"))
		return
	}

	_, err = s.getLatestSandbox(ctx, sandboxName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest sandbox release: %w", err))
		return
	}

	//// TODO(jm): create a new sandbox config
	//if _, err := s.updateAppSandbox(ctx, appID, sandboxRelease.ID); err != nil {
	//ctx.Error(fmt.Errorf("unable to update app sandbox: %w", err))
	//return
	//}

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
