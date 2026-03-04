package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID					GetQueueSignal
// @Summary			Get queue signal details
// @Description		Get detailed information about a specific queue signal
// @Tags				queues
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Param				queue_id	path		string	true	"Queue ID"
// @Param				signal_id	path		string	true	"Signal ID"
// @Success			200			{object}	app.QueueSignal
// @Failure			404			{object}	stderr.ErrResponse
// @Router				/v1/queues/{queue_id}/signals/{signal_id} [get]
func (s *service) GetQueueSignal(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")
	signalID := ctx.Param("signal_id")
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

	// Get the signal
	var signal app.QueueSignal
	res = s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Org").
		Preload("Queue").
		Preload("Emitter").
		Where("id = ?", signalID).
		Where("queue_id = ?", queueID).
		Where("org_id = ?", org.ID).
		First(&signal)

	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get signal %s: %w", signalID, res.Error))
		return
	}

	ctx.JSON(http.StatusOK, signal)
}
