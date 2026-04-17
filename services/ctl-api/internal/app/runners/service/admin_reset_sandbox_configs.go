package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminResetSandboxConfigs(ctx *gin.Context) {
	if res := s.db.WithContext(ctx).
		Where("job_type != ''").
		Delete(&app.SandboxModeJobConfig{}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to reset sandbox configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"reset": true})
}
