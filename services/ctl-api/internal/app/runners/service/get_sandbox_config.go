package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetRunnerSandboxConfig returns a single sandbox config for a job type and operation.
// Query params: job_type (required), operation (optional).
// Fallback: if no config matches the exact operation, returns the config with no operation set.
func (s *service) GetRunnerSandboxConfig(ctx *gin.Context) {
	jobType := ctx.Query("job_type")
	operation := ctx.Query("operation")

	if jobType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_type query param is required"})
		return
	}

	// Try exact match with operation first
	if operation != "" {
		var cfg app.SandboxModeJobConfig
		if res := s.db.WithContext(ctx).
			Where(app.SandboxModeJobConfig{JobType: jobType, Operation: operation}).
			First(&cfg); res.Error == nil {
			ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
			return
		}
	}

	// Fall back to job-type-only config (empty operation)
	var cfg app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeJobConfig{JobType: jobType, Operation: ""}).
		First(&cfg); res.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("no sandbox config for job_type=%s", jobType)})
		return
	}

	ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
}
