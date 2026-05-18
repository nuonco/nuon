package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) ReEnqueueSignal(c *gin.Context) {
	queueID := c.Param("id")
	signalID := c.Param("signal_id")

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

	if err := s.enqueuer.EnqueueInline(ctx, signalID, "admin-re-enqueue"); err != nil {
		s.l.Error("failed to re-enqueue signal",
			zap.Error(err),
			zap.String("queue_id", queueID),
			zap.String("signal_id", signalID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to re-enqueue signal"})
		return
	}

	s.l.Info("re-enqueued signal",
		zap.String("queue_id", queueID),
		zap.String("signal_id", signalID))

	c.JSON(http.StatusOK, gin.H{
		"status":          "re-enqueued",
		"queue_signal_id": signalID,
	})
}
