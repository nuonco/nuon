package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/apps
//
// Get all apps
//
//	@Summary	get all app installers for all orgs
//	@Schemes
//	@Description	get all apps
//	@Tags			apps/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	app.App
//	@Router			/v1/installers [get]
func (s *service) GetAllAppInstallers(ctx *gin.Context) {
	apps, err := s.getAllApps(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getAllAppInstallers(ctx context.Context) ([]*app.AppInstaller, error) {
	var apps []*app.AppInstaller
	res := s.db.WithContext(ctx).
		Preload("AppInstallerMetadata").
		Preload("App").
		Find(&apps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all apps: %w", res.Error)
	}
	return apps, nil
}
