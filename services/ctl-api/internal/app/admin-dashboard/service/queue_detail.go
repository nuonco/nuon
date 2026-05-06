package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

func (s *service) QueueDetail(c *gin.Context) {
	queueID := c.Param("id")

	var q app.Queue
	res := s.db.WithContext(c.Request.Context()).
		Preload("Emitters").
		Where("id = ?", queueID).
		First(&q)

	if res.Error != nil {
		s.l.Error("failed to fetch queue", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Queue not found"})
		return
	}

	// Get live status from Temporal (best effort)
	var status *queue.StatusResponse
	statusResp, err := s.getQueueStatusFromTemporal(c.Request.Context(), q.Workflow.ID)
	if err != nil {
		s.l.Warn("failed to get queue status from temporal", zap.Error(err), zap.String("queue_id", queueID))
	} else {
		status = statusResp
	}

	// Get recent signals
	var signals []app.QueueSignal
	s.db.WithContext(c.Request.Context()).
		Where("queue_id = ?", queueID).
		Order("created_at desc").
		Limit(20).
		Find(&signals)

	// Get in-flight signals for this queue
	var inFlightSignals []app.QueueSignal
	s.db.WithContext(c.Request.Context()).
		Where("queue_id = ?", queueID).
		Where("status->>'status' IN ('executing', 'in-progress', 'active')").
		Order("updated_at desc").
		Limit(50).
		Find(&inFlightSignals)

	c.JSON(http.StatusOK, gin.H{
		"queue":             &q,
		"status":            status,
		"signals":           signals,
		"in_flight_signals": inFlightSignals,
		"temporal_ui_url":   s.cfg.TemporalUIURL,
	})
}

func (s *service) QueueInFlightSignalsTable(c *gin.Context) {
	queueID := c.Param("id")

	var signals []app.QueueSignal
	s.db.WithContext(c.Request.Context()).
		Where("queue_id = ?", queueID).
		Where("status->>'status' IN ('executing', 'in-progress', 'active')").
		Order("updated_at desc").
		Limit(50).
		Find(&signals)

	c.JSON(http.StatusOK, gin.H{
		"signals":  signals,
		"queue_id": queueID,
	})
}

func (s *service) getQueueStatusFromTemporal(ctx context.Context, workflowID string) (*queue.StatusResponse, error) {
	// Use the update-with-start approach that the queue client uses
	encodedValue, err := s.temporalClient.QueryWorkflow(ctx, workflowID, "", queue.StatusHandlerName)
	if err != nil {
		return nil, fmt.Errorf("unable to query workflow: %w", err)
	}

	var statusResp queue.StatusResponse
	if err := encodedValue.Get(&statusResp); err != nil {
		return nil, fmt.Errorf("unable to decode workflow status: %w", err)
	}

	return &statusResp, nil
}
