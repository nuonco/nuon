package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminDeleteSandboxSignalConfig(ctx *gin.Context) {
	signalType := ctx.Param("signal_type")

	if res := s.db.WithContext(ctx).
		Where(app.SandboxSignalConfig{SignalType: signalType}).
		Delete(&app.SandboxSignalConfig{}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete sandbox signal config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
