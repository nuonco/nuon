package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/apps/installs
// Get an app's installs
// @Summary get all installs for an app
// @Schemes
// @Description get all installs for an org
// @Param app_id path string app_id "app ID"
// @Tags installs
// @Accept json
// @Produce json
// @Success 201 {object} app.Install
// @Router /v1/apps/{app_id}/installs [GET]
func (s *service) GetAppInstalls(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	if appID == "" {
		ctx.Error(fmt.Errorf("app id must be passed in"))
		return
	}

	install, err := s.getAppInstalls(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) getAppInstalls(ctx context.Context, appID string) ([]app.Install, error) {
	currentApp := &app.App{}
	res := s.db.WithContext(ctx).
		Preload("Installs").
		Preload("Installs.SandboxRelease").
		Preload("Installs.AWSAccount").
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return currentApp.Installs, nil
}
