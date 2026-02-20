package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) AddSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.Error(err))
		ctx.String(500, "Failed to get organization")
		return
	}

	results, err := s.orgsHelpers.AddSupportUsersToOrg(ctx, org)
	if err != nil {
		s.l.Error("failed to add support users", zap.Error(err))
		ctx.String(500, "Failed to add support users")
		return
	}

	successCount := 0
	alreadyExistsCount := 0
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
			s.l.Warn("failed to add support user",
				zap.String("email", result.Email),
				zap.Error(result.Error))
		} else if result.AlreadyExists {
			alreadyExistsCount++
		} else if result.Success {
			successCount++
		}
	}

	s.l.Info("support users operation completed",
		zap.String("org_id", orgID),
		zap.Int("success", successCount),
		zap.Int("already_exists", alreadyExistsCount),
		zap.Int("errors", errorCount))

	// Render toast component
	component := views.SupportUsersToast(successCount, alreadyExistsCount, errorCount)

	// Log that we're about to render
	s.l.Debug("rendering support users toast")

	ctx.Header("Content-Type", "text/html; charset=utf-8")

	if err := component.Render(ctx.Request.Context(), ctx.Writer); err != nil {
		s.l.Error("failed to render toast", zap.Error(err))
		ctx.String(500, "Failed to render response")
		return
	}

	s.l.Debug("toast rendered successfully")
}
