package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) RestartQueue(c *gin.Context) {
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

	if err := s.queueClient.Restart(c.Request.Context(), q.ID); err != nil {
		s.l.Error("failed to restart queue", zap.Error(err), zap.String("queue_id", queueID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart queue"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/queues/"+queueID)
}
