package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// OrgQueueSignals returns all queue signals for an org.
func (s *service) OrgQueueSignals(c *gin.Context) {
	orgID := c.Param("id")

	var signals []app.QueueSignal
	if res := s.readDB().WithContext(c.Request.Context()).
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Limit(200).
		Find(&signals); res.Error != nil {
		s.l.Error("failed to list org queue signals", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list queue signals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signals": signals})
}

// DeleteOrgQueueSignals hard-deletes all queue signals for an org.
func (s *service) DeleteOrgQueueSignals(c *gin.Context) {
	orgID := c.Param("id")

	res := s.db.WithContext(c.Request.Context()).
		Where("org_id = ?", orgID).
		Unscoped().
		Delete(&app.QueueSignal{})
	if res.Error != nil {
		s.l.Error("failed to delete org queue signals", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete queue signals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "deleted",
		"signals_deleted": res.RowsAffected,
	})
}
