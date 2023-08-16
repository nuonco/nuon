package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/installs
// Get an install component
// @Summary get an install component
// @Schemes
// @Description get an install
// @Param install_id path string install_id "install ID"
// @Param component_id path string install_id "component ID"
// @Tags installs
// @Accept json
// @Produce json
// @Success 200 {object} app.InstallComponent
// @Router /v1/installs/{install_id}/component/{component_id} [get]
func (s *service) GetInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	if installID == "" {
		ctx.Error(fmt.Errorf("install_id must be passed in"))
		return
	}
	componentID := ctx.Param("component_id")
	if componentID == "" {
		ctx.Error(fmt.Errorf("component_id must be passed in"))
		return
	}

	installCmp, err := s.getInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  install cmp %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, installCmp)
}

func (s *service) getInstallComponent(ctx context.Context, installID, componentID string) (*app.Install, error) {
	return nil, nil
}
