package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
)

type RestartAppRequest struct{}

//	@ID						AdminRestartApp
//	@Summary				restart an apps event loop
//	@Description.markdown	restart_app.md
//	@Param					app_id	path	string				true	"app ID"
//	@Param					req		body	RestartAppRequest	true	"Input"
//	@Tags					apps/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{boolean}	true
//	@Router					/v1/apps/{app_id}/admin-restart [POST]
func (s *service) RestartApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req RestartAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	app, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	s.evClient.Send(ctx, app.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) getApp(ctx context.Context, appID string) (*app.App, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Components").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Where("name = ?", appID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
