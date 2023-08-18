package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @BasePath /v1/apps/installs
// Get an install component deploys
// @Summary get an install components deploys
// @Schemes
// @Description get all deploys for an install
// @Param install_id path string true "install ID"
// @Param component_id path string true "component ID"
// @Tags installs
// @Accept json
// @Produce json
// @Success 201 {array} app.InstallDeploy
// @Router /v1/installs/{install_id}/components/{component_id}/deploys [GET]
func (s *service) GetInstallComponentDeploys(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installComponentDeploys, err := s.getInstallComponentDeploys(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component deploys: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installComponentDeploys)
}

func (s *service) getInstallComponentDeploys(ctx context.Context, installID, componentID string) ([]app.InstallDeploy, error) {
	install := app.InstallComponent{
		ComponentID: componentID,
		InstallID:   installID,
	}
	res := s.db.WithContext(ctx).Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
		return db.Order("install_deploys.created_at DESC").Limit(1000)
	}).
		Preload("InstallDeploys.Build").
		Preload("InstallDeploys.Build.VCSConnectionCommit").
		Where(install).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return install.InstallDeploys, nil
}
