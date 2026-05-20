package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) CheckCANQueue(c *gin.Context) {
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

	resp, err := s.queueClient.CheckCAN(ctx, q.ID)
	if err != nil {
		s.l.Error("failed to check CAN", zap.Error(err), zap.String("queue_id", queueID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check CAN"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
