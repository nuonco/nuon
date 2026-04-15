package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminResetSandboxConfigs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	if res := s.db.WithContext(ctx).
		Where(app.SandboxConfig{RunnerID: runnerID}).
		Delete(&app.SandboxConfig{}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to reset sandbox configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"reset": true})
}
