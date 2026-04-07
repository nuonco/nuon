package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	defaultAwaitTimeout = 30 * time.Second
	maxAwaitTimeout     = 120 * time.Second
)

// @ID					AwaitQueueSignal
// @Summary			Long-poll for queue signal completion
// @Description		Blocks until the queue signal finishes or the timeout is reached
// @Tags				queues
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Param				queue_id	path		string	true	"Queue ID"
// @Param				signal_id	path		string	true	"Signal ID"
// @Param				timeout		query		int		false	"Timeout in seconds (default 30, max 120)"
// @Success			200			{object}	app.QueueSignal
// @Failure			408			{object}	stderr.ErrResponse
// @Failure			404			{object}	stderr.ErrResponse
// @Router				/v1/queues/{queue_id}/signals/{signal_id}/await [get]
func (s *service) AwaitQueueSignal(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")
	signalID := ctx.Param("signal_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	timeout := defaultAwaitTimeout
	if t := ctx.Query("timeout"); t != "" {
		secs, err := strconv.Atoi(t)
		if err != nil || secs <= 0 {
			ctx.Error(fmt.Errorf("invalid timeout: %s", t))
			return
		}
		timeout = time.Duration(secs) * time.Second
		if timeout > maxAwaitTimeout {
			timeout = maxAwaitTimeout
		}
	}

	// Verify queue exists and user has access
	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where("id = ?", queueID).
		Where("org_id = ?", org.ID).
		First(&queue); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get queue %s: %w", queueID, res.Error))
		return
	}

	// Verify signal exists and belongs to queue
	var signal app.QueueSignal
	if res := s.db.WithContext(ctx).
		Where("id = ?", signalID).
		Where("queue_id = ?", queueID).
		Where("org_id = ?", org.ID).
		First(&signal); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get signal %s: %w", signalID, res.Error))
		return
	}

	awaitCtx, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
	defer cancel()

	_, err = s.queueClient.AwaitSignal(awaitCtx, signalID)
	if err != nil {
		if awaitCtx.Err() != nil {
			ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout waiting for signal completion"})
			return
		}
		ctx.Error(fmt.Errorf("error awaiting signal: %w", err))
		return
	}

	// Fetch the final signal state
	if res := s.db.WithContext(ctx).
		Where("id = ?", signalID).
		Where("queue_id = ?", queueID).
		Where("org_id = ?", org.ID).
		First(&signal); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get signal %s: %w", signalID, res.Error))
		return
	}

	ctx.JSON(http.StatusOK, signal)
}
