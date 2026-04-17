package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminDeleteSandboxConfig(ctx *gin.Context) {
	configID := ctx.Param("config_id")

	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeJobConfig{ID: configID}).
		Delete(&app.SandboxModeJobConfig{}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete sandbox config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
