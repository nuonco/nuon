package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminListAllSandboxConfigs(ctx *gin.Context) {
	var configs []app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Order("job_type asc").
		Find(&configs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list sandbox configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}
