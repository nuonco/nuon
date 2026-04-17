package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminGetSandboxConfigs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	configs, err := s.getSandboxConfigs(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}

func (s *service) getSandboxConfigs(ctx context.Context, runnerID string) ([]app.SandboxModeJobConfig, error) {
	var configs []app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where("job_type != ''").
		Order("job_type asc").
		Find(&configs); res.Error != nil {
		return nil, fmt.Errorf("unable to find sandbox configs: %w", res.Error)
	}

	return configs, nil
}
