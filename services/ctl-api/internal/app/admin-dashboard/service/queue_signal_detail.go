package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) QueueSignalDetail(c *gin.Context) {
	queueID := c.Param("id")
	signalID := c.Param("signal_id")

	var signal app.QueueSignal
	res := s.db.WithContext(c.Request.Context()).
		Preload("Emitter").
		Where("id = ? AND queue_id = ?", signalID, queueID).
		First(&signal)

	if res.Error != nil {
		s.l.Error("failed to fetch queue signal", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
		return
	}

	// Fetch the parent queue for context
	var q app.Queue
	s.db.WithContext(c.Request.Context()).
		Where("id = ?", queueID).
		First(&q)

	component := views.QueueSignalDetail(&signal, &q, s.cfg.TemporalUIURL)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
