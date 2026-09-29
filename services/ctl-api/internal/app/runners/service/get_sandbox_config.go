package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) GetRunnerSandboxConfig(ctx *gin.Context) {
	jobType := ctx.Query("job_type")
	operation := ctx.Query("operation")

	if jobType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_type query param is required"})
		return
	}

	if operation != "" {
		var cfg app.SandboxModeJobConfig
		if res := s.db.WithContext(ctx).
			Where(app.SandboxModeJobConfig{JobType: jobType, Operation: operation, Enabled: true}).
			First(&cfg); res.Error == nil {
			ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
			return
		}
	}

	var cfg app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeJobConfig{JobType: jobType, Operation: "all", Enabled: true}).
		First(&cfg); res.Error == nil {
		ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
		return
	}

	// why: Fall back to job-type-only config (empty operation).
	// NOTE: must use map-based Where because GORM silently drops zero-value
	// fields from struct-based Where, and "" is a zero value for string.
	if res := s.db.WithContext(ctx).
		Where(map[string]interface{}{"job_type": jobType, "operation": "", "enabled": true}).
		First(&cfg); res.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("no sandbox config for job_type=%s", jobType)})
		return
	}

	ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
}
