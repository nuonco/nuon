package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
)

// @ID        CreateAppSandboxBuild
// @Summary   create app sandbox build
// @Tags      apps
// @Accept    json
// @Produce   json
// @Param     app_id  path  string  true  "app ID"
// @Security  APIKey
// @Security  OrgID
// @Failure   400  {object}  stderr.ErrResponse
// @Failure   401  {object}  stderr.ErrResponse
// @Failure   404  {object}  stderr.ErrResponse
// @Failure   500  {object}  stderr.ErrResponse
// @Success   201  {object}  app.AppSandboxBuild
// @Router    /v1/apps/{app_id}/sandbox/builds [post]
func (s *service) CreateAppSandboxBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	currentApp, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	// get latest app config
	var latestConfig app.AppConfig
	res := s.db.WithContext(ctx).
		Where("app_id = ?", currentApp.ID).
		Order("created_at DESC").
		First(&latestConfig)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("no app config found for app %s: %w", appID, gorm.ErrRecordNotFound))
		return
	}

	// get latest sandbox config
	var sandboxConfig app.AppSandboxConfig
	res = s.db.WithContext(ctx).
		Where("app_id = ?", currentApp.ID).
		Order("created_at DESC").
		First(&sandboxConfig)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("no sandbox config found for app %s: %w", appID, gorm.ErrRecordNotFound))
		return
	}

	// create the sandbox build record
	build := app.AppSandboxBuild{
		AppID:              currentApp.ID,
		AppConfigID:        latestConfig.ID,
		AppSandboxConfigID: sandboxConfig.ID,
		Status:             app.AppSandboxBuildStatusQueued,
		StatusDescription:  "queued and waiting for runner",
	}
	if res := s.db.WithContext(ctx).Create(&build); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create sandbox build: %w", res.Error))
		return
	}

	// send signal to trigger the build workflow
	s.evClient.Send(ctx, currentApp.ID, &signals.Signal{
		Type:              signals.OperationBuildSandbox,
		AppSandboxBuildID: build.ID,
	})

	ctx.JSON(http.StatusCreated, build)
}
