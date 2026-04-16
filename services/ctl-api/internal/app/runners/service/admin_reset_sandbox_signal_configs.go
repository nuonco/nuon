package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminResetSandboxSignalConfigs(ctx *gin.Context) {
	if res := s.db.WithContext(ctx).
		Where("1 = 1").
		Delete(&app.SandboxModeSignalConfig{}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to reset sandbox signal configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
