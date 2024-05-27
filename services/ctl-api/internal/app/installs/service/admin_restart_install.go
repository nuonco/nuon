package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type RestartInstallRequest struct{}

// @ID AdminRestartInstall
// @Summary	restart an installs event loop
// @Description.markdown restart_install.md
// @Param			install_id	path	string					true	"install ID"
// @Param			req			body	RestartInstallRequest	true	"Input"
// @Tags			installs/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/installs/{install_id}/admin-restart [POST]
func (s *service) RestartInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req RestartInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("App").
		Preload("App.AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		Preload("App.Org").
		Preload("AppSandboxConfig").
		Where("name = ?", installID).
		Or("id = ?", installID).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
