package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminUpdateAppSandboxConfigsRequest struct{}

//	@BasePath	/v1/apps
//
// Update app sandbox configs
//
//	@Summary	update all app sandbox configs
//
// @Schemes
//
//	@Description	update and app to the latest sandbox
//	@Param			req		body	AdminUpdateSandboxRequest	true	"Input"
//	@Tags			apps/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/apps/admin-update-app-sandbox-configs [POST]
func (s *service) AdminUpdateAppSandboxConfigs(ctx *gin.Context) {
	var req AdminUpdateAppSandboxConfigsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	cfgs, err := s.getAllAppSandboxConfigs(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app sandbox configs: %w", err))
		return
	}

	for _, cfg := range cfgs {
		appSandboxCfg := app.AppSandboxConfig{
			ID: cfg.ID,
		}
		res := s.db.WithContext(ctx).
			Model(&appSandboxCfg).
			Update("app_id", cfg.AppID)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to update app sandbox config: %w", res.Error))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getAllAppSandboxConfigs(ctx context.Context) ([]*app.AppSandboxConfig, error) {
	var cfgs []*app.AppSandboxConfig

	res := s.db.Unscoped().
		WithContext(ctx).
		First(&cfgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app sandbox configs: %w", res.Error)
	}

	return cfgs, nil
}
