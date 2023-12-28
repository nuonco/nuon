package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetInstallComponents
// @Summary	get an installs components
// @Description.markdown	get_install_components.md
// @Param			install_id	path	string	true	"install ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.InstallComponent
// @Router			/v1/installs/{install_id}/components [GET]
func (s *service) GetInstallComponents(ctx *gin.Context) {
	appID := ctx.Param("install_id")
	installComponents, err := s.getInstallComponents(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installComponents)
}

func (s *service) getInstallComponents(ctx context.Context, installID string) ([]app.InstallComponent, error) {
	install := &app.Install{}
	res := s.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_components.created_at DESC")
		}).
		Preload("InstallComponents.InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC")
		}).
		Preload("InstallComponents.Component").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	return install.InstallComponents, nil
}
