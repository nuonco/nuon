package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

// @ID						GetInstallLatestDeploy
// @Summary				get an install's latest deploy
// @Description.markdown	get_install_latest_deploy.md
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallDeploy
// @Router					/v1/installs/{install_id}/deploys/latest [get]
func (s *service) GetInstallLatestDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	installDeploy, err := s.getInstallLatestDeploy(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install latest deploy: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploy)
}

func (s *service) getInstallLatestDeploy(ctx context.Context, installID string) (*app.InstallDeploy, error) {
	installCmp := &app.InstallComponent{}
	res := s.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC").Limit(1000)
		}).
		Preload("TerraformWorkspace").
		First(&installCmp, "install_id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}
	if len(installCmp.InstallDeploys) != 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("no deploy exists for install: %w", gorm.ErrRecordNotFound),
			Description: "no errors exist for install yet",
		}
	}

	return &installCmp.InstallDeploys[0], nil
}
