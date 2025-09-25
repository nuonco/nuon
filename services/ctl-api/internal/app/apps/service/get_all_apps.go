package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						GetAllApps
//	@Summary				get all apps for all orgs
//	@Description.markdown	get_all_apps.md
//	@Tags					apps/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{array}	app.App
//	@Router					/v1/apps [get]
func (s *service) GetAllApps(ctx *gin.Context) {
	apps, err := s.getAllApps(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getAllApps(ctx context.Context) ([]*app.App, error) {
	var apps []*app.App
	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfigs").
		Preload("Installs").
		Preload("Components").
		Order("created_at desc").
		Find(&apps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all apps: %w", res.Error)
	}
	return apps, nil
}
