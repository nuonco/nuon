package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/apps
// Get all apps for the current org
// @Summary get all apps for the current org
// @Schemes
// @Description get an app
// @Tags apps
// @Accept json
// @Produce json
// @Success 200 {array} app.App
// @Router /v1/apps [get]
func (s *service) GetApps(ctx *gin.Context) {
	orgID := ctx.Param("app_id")

	apps, err := s.getApps(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for %s: %w", orgID, err))
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getApps(ctx context.Context, orgID string) ([]*app.App, error) {
	var apps []*app.App
	org := &app.Org{
		Model: app.Model{ID: orgID},
	}

	res := s.db.WithContext(ctx).Model(&org).Association("Apps").Find(&apps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org apps: %w", res.Error)
	}

	return apps, nil
}
