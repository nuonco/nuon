package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartAppRequest struct{}

// @ID						AdminRestartApp
// @Summary				restart an apps event loop
// @Description.markdown	restart_app.md
// @Param					app_id	path	string				true	"app ID"
// @Param					req		body	RestartAppRequest	true	"Input"
// @Tags					apps/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/apps/{app_id}/admin-restart [POST]
func (s *service) RestartApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req RestartAppRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	_, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

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
