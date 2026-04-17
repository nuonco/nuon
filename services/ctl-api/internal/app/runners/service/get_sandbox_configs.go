package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetRunnerSandboxConfigs returns sandbox configs for the authenticated runner.
func (s *service) GetRunnerSandboxConfigs(ctx *gin.Context) {
	var configs []app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where("job_type != ''").
		Order("job_type asc").
		Find(&configs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox configs: %w", res.Error))
		return
	}

	resp := make([]SandboxConfigResponse, 0, len(configs))
	for _, cfg := range configs {
		resp = append(resp, convertToSandboxConfigResponse(cfg))
	}

	ctx.JSON(http.StatusOK, resp)
}
