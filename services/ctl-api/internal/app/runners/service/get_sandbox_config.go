package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetRunnerSandboxConfig returns a single sandbox config for a job type and operation.
// Query params: job_type (required), operation (optional).
// Lookup order:
//  1. Exact match on (job_type, operation)
//  2. Wildcard match on (job_type, "all") — matches any operation
//  3. Legacy fallback on (job_type, "") — empty operation
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

	// Fall back to "all" operation config (wildcard for any operation)
	var cfg app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeJobConfig{JobType: jobType, Operation: "all"}).
		First(&cfg); res.Error == nil {
		ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
		return
	}

	// Fall back to job-type-only config (empty operation).
	// NOTE: must use map-based Where because GORM silently drops zero-value
	// fields from struct-based Where, and "" is a zero value for string.
	if res := s.db.WithContext(ctx).
		Where(map[string]interface{}{"job_type": jobType, "operation": ""}).
		First(&cfg); res.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("no sandbox config for job_type=%s", jobType)})
		return
	}

	ctx.JSON(http.StatusOK, convertToSandboxConfigResponse(cfg))
}
