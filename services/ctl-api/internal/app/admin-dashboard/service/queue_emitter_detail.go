package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) QueueEmitterDetail(c *gin.Context) {
	queueID := c.Param("id")
	emitterID := c.Param("emitter_id")

	var emitter app.QueueEmitter
	res := s.db.WithContext(c.Request.Context()).
		Where("id = ? AND queue_id = ?", emitterID, queueID).
		First(&emitter)

	if res.Error != nil {
		s.l.Error("failed to fetch queue emitter", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Emitter not found"})
		return
	}

	// Fetch the parent queue for context
	var q app.Queue
	s.db.WithContext(c.Request.Context()).
		Where("id = ?", queueID).
		First(&q)

	// Fetch signals created by this emitter
	var signals []app.QueueSignal
	s.db.WithContext(c.Request.Context()).
		Where("emitter_id = ?", emitterID).
		Order("created_at desc").
		Limit(50).
		Find(&signals)

	component := views.QueueEmitterDetail(&emitter, &q, signals, s.cfg.TemporalUIURL)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
