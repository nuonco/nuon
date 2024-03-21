package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetInstallerInstall
// @Summary	render an installer install
// @Description.markdown	get_installer_install.md
// @Tags installers
// @Accept			json
// @Produce		json
// @Param			installer_slug	path		string	true	"installer slug or ID"
// @Param			install_id		path		string	true	"install id"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.Install
// @Router			/v1/installer/{installer_slug}/install/{install_id} [get]
func (s *service) GetInstallerInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) findInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		Preload("AppRunnerConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs.created_at DESC")
		}).
		Preload("App").
		Preload("App.Org").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
