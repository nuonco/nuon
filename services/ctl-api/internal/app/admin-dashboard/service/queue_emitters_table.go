package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) QueueEmittersTable(c *gin.Context) {
	queueID := c.Param("id")

	var emitters []app.QueueEmitter

	res := s.db.WithContext(c.Request.Context()).
		Where("queue_id = ?", queueID).
		Order("created_at desc").
		Find(&emitters)

	if res.Error != nil {
		s.l.Error("failed to fetch emitters", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emitters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emitters": emitters,
		"queue_id": queueID,
	})
}
