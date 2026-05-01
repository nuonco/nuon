package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) AddSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization"})
		return
	}

	results, err := s.orgsHelpers.AddSupportUsersToOrg(ctx, org)
	if err != nil {
		s.l.Error("failed to add support users", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add support users"})
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

	ctx.JSON(http.StatusOK, gin.H{
		"success":        successCount,
		"already_exists": alreadyExistsCount,
		"errors":         errorCount,
	})
}
