package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetInstallComponentLatestDeploy
// @Summary	get the latest deploy for an install component
// @Description.markdown	get_install_component_latest_deploy.md
// @Param			install_id		path	string	true	"install ID"
// @Param			component_id	path	string	true	"component ID"
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
// @Success		200				{object}	app.InstallDeploy
// @Router			/v1/installs/{install_id}/components/{component_id}/deploys/latest [get]
func (s *service) GetInstallComponentLatestDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installDeploy, err := s.getInstallComponentLatestDeploy(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component latest deploy: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploy)
}

func (s *service) getInstallComponentLatestDeploy(ctx context.Context, installID string, componentID string) (*app.InstallDeploy, error) {
	installCmp := app.InstallComponent{}
	res := s.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC").Limit(1000)
		}).
		Where(&app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installCmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}
	if len(installCmp.InstallDeploys) != 1 {
		return nil, fmt.Errorf("no deploy exists for install: %w", gorm.ErrRecordNotFound)
	}

	return &installCmp.InstallDeploys[0], nil
}
