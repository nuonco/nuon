package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID AdminGetInstall
// @Summary	get an install
// @Description.markdown get_install.md
// @Param			install_id	path	string					true	"install ID"
// @Tags			installs/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/installs/{install_id}/admin-get [GET]
func (s *service) AdminGetInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.adminGetInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) adminGetInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("App").
		Preload("App.Org").
		Preload("CreatedBy").
		Preload("InstallInputs").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").
		Or("id = ?", installID).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
