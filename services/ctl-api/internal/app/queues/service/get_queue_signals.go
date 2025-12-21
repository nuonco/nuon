package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID					GetQueueSignals
// @Summary			List queue signals
// @Description		Get a list of signals for a specific queue with optional filtering
// @Tags				queues
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Param				queue_id	path		string	true	"Queue ID"
// @Param				owner_id	query		string	false	"Filter by owner ID"
// @Param				owner_type	query		string	false	"Filter by owner type (e.g., app_branches)"
// @Param				status		query		string	false	"Filter by status"
// @Param				type		query		string	false	"Filter by signal type"
// @Param				limit		query		int		false	"Limit results (default: 50)"
// @Param				offset		query		int		false	"Offset for pagination (default: 0)"
// @Success			200			{array}		app.QueueSignal
// @Failure			404			{object}	stderr.ErrResponse
// @Router				/v1/queues/{queue_id}/signals [get]
func (s *service) GetQueueSignals(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Verify queue exists and user has access
	var queue app.Queue
	res := s.db.WithContext(ctx).
		Where("id = ?", queueID).
		Where("org_id = ?", org.ID).
		First(&queue)

	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get queue %s: %w", queueID, res.Error))
		return
	}

	// Parse query parameters
	ownerID := ctx.Query("owner_id")
	ownerType := ctx.Query("owner_type")
	status := ctx.Query("status")
	signalType := ctx.Query("type")

	limit := 50
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build query with filters
	query := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Org").
		Preload("Queue").
		Preload("Emitter").
		Where("queue_id = ?", queueID).
		Where("org_id = ?", org.ID)

	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	if ownerType != "" {
		query = query.Where("owner_type = ?", ownerType)
	}
	if status != "" {
		query = query.Where("status->>'status' = ?", status)
	}
	if signalType != "" {
		query = query.Where("type = ?", signalType)
	}

	var signals []app.QueueSignal
	res = query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&signals)

	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get queue signals: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, signals)
}
