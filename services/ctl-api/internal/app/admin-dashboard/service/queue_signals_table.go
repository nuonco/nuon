package service

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const signalsPerPage = 20

func (s *service) QueueSignalsTable(c *gin.Context) {
	queueID := c.Param("id")
	page := getPageFromQuery(c)

	var signals []app.QueueSignal
	var totalCount int64

	query := s.readDB().WithContext(c.Request.Context()).
		Model(&app.QueueSignal{}).
		Where("queue_id = ?", queueID)

	if err := query.Count(&totalCount).Error; err != nil {
		s.l.Error("failed to count signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count signals"})
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(signalsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * signalsPerPage

	res := s.readDB().WithContext(c.Request.Context()).
		Where("queue_id = ?", queueID).
		Order("created_at desc").
		Limit(signalsPerPage).
		Offset(offset).
		Find(&signals)

	if res.Error != nil {
		s.l.Error("failed to fetch signals", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch signals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signals":     signals,
		"queue_id":    queueID,
		"page":        page,
		"total_pages": totalPages,
	})
}
