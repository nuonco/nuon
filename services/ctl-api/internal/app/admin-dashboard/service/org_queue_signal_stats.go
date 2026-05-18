package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type queueSignalStatRow struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// OrgQueueSignalStats returns queue signal counts grouped by type and status for an org.
func (s *service) OrgQueueSignalStats(c *gin.Context) {
	orgID := c.Param("id")

	var rows []queueSignalStatRow
	if res := s.readDB().WithContext(c.Request.Context()).
		Table("queue_signals").
		Select("type, status->>'status' as status, COUNT(*) as count").
		Where("org_id = ? AND deleted_at = 0", orgID).
		Group("type, status->>'status'").
		Order("count DESC").
		Scan(&rows); res.Error != nil {
		s.l.Error("failed to query queue signal stats", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query queue signal stats"})
		return
	}

	total := int64(0)
	for _, r := range rows {
		total += r.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": rows,
		"total": total,
	})
}
