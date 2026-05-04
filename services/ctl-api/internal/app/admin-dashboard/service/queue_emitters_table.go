package service

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const queueEmittersPerPage = 20

func (s *service) QueueEmittersTable(c *gin.Context) {
	ctx := c.Request.Context()
	queueID := c.Param("id")
	page := getPageFromQuery(c)

	query := s.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("queue_id = ?", queueID)

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.l.Error("failed to count emitters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emitters"})
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(queueEmittersPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * queueEmittersPerPage

	var emitters []app.QueueEmitter
	if err := query.
		Order("created_at desc").
		Limit(queueEmittersPerPage).
		Offset(offset).
		Find(&emitters).Error; err != nil {
		s.l.Error("failed to fetch emitters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emitters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emitters":    emitters,
		"queue_id":    queueID,
		"page":        page,
		"total_pages": totalPages,
	})
}
