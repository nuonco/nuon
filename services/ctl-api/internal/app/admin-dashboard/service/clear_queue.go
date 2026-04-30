package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) ClearQueue(c *gin.Context) {
	queueID := c.Param("id")

	var q app.Queue
	res := s.db.WithContext(c.Request.Context()).
		Where("id = ?", queueID).
		First(&q)

	if res.Error != nil {
		s.l.Error("failed to fetch queue", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Queue not found"})
		return
	}

	ctx := c.Request.Context()
	if q.OrgID != nil {
		ctx = cctx.SetOrgIDContext(ctx, *q.OrgID)
	}

	if err := s.queueClient.ClearQueue(ctx, q.ID); err != nil {
		s.l.Error("failed to clear queue", zap.Error(err), zap.String("queue_id", queueID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cleared"})
}
