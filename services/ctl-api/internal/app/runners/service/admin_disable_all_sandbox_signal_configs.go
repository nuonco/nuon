package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminDisableAllSandboxSignalConfigs(ctx *gin.Context) {
	if res := s.db.WithContext(ctx).
		Model(&app.SandboxModeSignalConfig{}).
		Where("enabled = ?", true).
		Update("enabled", false); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to disable all sandbox signal configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
