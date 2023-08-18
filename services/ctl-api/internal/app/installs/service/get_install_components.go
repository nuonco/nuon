package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/apps/installs
// Get an install's components
// @Summary get an installs components
// @Schemes
// @Description get all components for an install
// @Param install_id path string true "install ID"
// @Tags installs
// @Accept json
// @Produce json
// @Success 201 {array} app.InstallComponent
// @Router /v1/installs/{install_id}/components [GET]
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
	res := s.db.WithContext(ctx).Preload("InstallComponents").First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return install.InstallComponents, nil
}
