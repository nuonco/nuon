package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

//	@BasePath	/v1/apps/installs
//
// Get an install's deploys
//
//	@Summary	get an installs deploys
//	@Schemes
//	@Description	get all deploys for an install
//	@Param			install_id	path	string	true	"install ID"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{array}		app.InstallDeploy
//	@Router			/v1/installs/{install_id}/deploys [GET]
func (s *service) GetInstallDeploys(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installDeploys, err := s.getInstallDeploys(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploys: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploys)
}

func (s *service) getInstallDeploys(ctx context.Context, installID string) ([]app.InstallDeploy, error) {
	var installCmps []app.InstallComponent
	res := s.db.WithContext(ctx).Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
		return db.Order("install_deploys.created_at DESC").Limit(1000)
	}).
		Preload("InstallDeploys.ComponentBuild").
		Preload("InstallDeploys.ComponentBuild.VCSConnectionCommit").
		Where("install_id = ?", installID).
		First(&installCmps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	deploys := make([]app.InstallDeploy, 0)
	for _, installCmp := range installCmps {
		deploys = append(deploys, installCmp.InstallDeploys...)
	}

	return deploys, nil
}
