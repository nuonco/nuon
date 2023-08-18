package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeployLog struct{}

// @BasePath /v1/installs
// Get install deploy logs
// @Summary get install deploy logs
// @Schemes
// @Description get install deploy logs
// @Param install_id path string true "install ID"
// @Param deploy_id path string true "deploy ID"
// @Tags installs
// @Accept json
// @Produce json
// @Success 200 {object} []DeployLog
// @Router /v1/installs/{install_id}/deploys/{deploy_id}/logs [get]
func (s *service) GetInstallDeployLogs(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	if installID == "" {
		ctx.Error(fmt.Errorf("install_id must be passed in"))
		return
	}
	deployID := ctx.Param("deploy_id")
	if deployID == "" {
		ctx.Error(fmt.Errorf("deploy_id must be passed in"))
		return
	}

	logs, err := s.getInstallDeployLogs(ctx, installID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy %s: %w", deployID, err))
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getInstallDeployLogs(ctx context.Context, installID, componentID string) ([]DeployLog, error) {
	return nil, nil
}
