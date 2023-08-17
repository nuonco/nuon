package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/installs
// Get an install deploy
// @Summary get an install deploy
// @Schemes
// @Description get an install deploy
// @Param install_id path string install_id "install ID"
// @Param deploy_id path string deploy_id "deploy ID"
// @Tags installs
// @Accept json
// @Produce json
// @Success 200 {object} app.InstallDeploy
// @Router /v1/installs/{install_id}/deploys/{deploy_id} [get]
func (s *service) GetInstallDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	deployID := ctx.Param("deploy_id")

	installDeploy, err := s.getInstallDeploy(ctx, installID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy %s: %w", deployID, err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploy)
}

func (s *service) getInstallDeploy(ctx context.Context, installID, deployID string) (*app.InstallDeploy, error) {
	installCmp := &app.InstallComponent{}
	res := s.db.WithContext(ctx).
		Preload("InstallDeploys", "id = ?", deployID).
		Preload("InstallDeploys.Build", "id = ?", deployID).
		First(&installCmp, "install_id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}
	if len(installCmp.InstallDeploys) != 1 {
		return nil, fmt.Errorf("deploy not found")
	}

	return &installCmp.InstallDeploys[0], nil
}
