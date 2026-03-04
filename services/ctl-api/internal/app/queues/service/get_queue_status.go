package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @ID					GetQueueStatus
// @Summary			Get live queue status
// @Description		Get real-time status of a queue including depth and in-flight signals
// @Tags				queues
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Param				queue_id	path		string	true	"Queue ID"
// @Success			200			{object}	queue.StatusResponse
// @Failure			404			{object}	stderr.ErrResponse
// @Router				/v1/queues/{queue_id}/status [get]
func (s *service) GetQueueStatus(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Verify queue exists and user has access
	var q app.Queue
	res := s.db.WithContext(ctx).
		Where("id = ?", queueID).
		Where("org_id = ?", org.ID).
		First(&q)

	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get queue %s: %w", queueID, res.Error))
		return
	}

	// Get queue status from Temporal workflow
	statusResp, err := s.getQueueStatusFromTemporal(ctx, q.Workflow.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get queue status from temporal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, statusResp)
}

func (s *service) getQueueStatusFromTemporal(ctx context.Context, workflowID string) (*queue.StatusResponse, error) {
	// Query the queue workflow for its status
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
