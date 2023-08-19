package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/apps
// Delete an app
// @Summary delete an app
// @Schemes
// @Description delete an app
// @Param app_id path string true "app ID"
// @Tags apps
// @Accept json
// @Produce json
// @Success 200 {boolean} true
// @Router /v1/apps/{app_id} [DELETE]
func (s *service) DeleteApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	err := s.deleteApp(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, appID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteApp(ctx context.Context, appID string) error {
	res := s.db.WithContext(ctx).Delete(&app.App{
		ID: appID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete app: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("app not found")
	}

	return nil
}
