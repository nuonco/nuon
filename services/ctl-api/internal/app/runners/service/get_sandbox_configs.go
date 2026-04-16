package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetRunnerSandboxConfigs returns sandbox configs for the authenticated runner.
func (s *service) GetRunnerSandboxConfigs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var configs []app.SandboxModeConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeConfig{RunnerID: runnerID}).
		Order("job_type asc").
		Find(&configs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}
