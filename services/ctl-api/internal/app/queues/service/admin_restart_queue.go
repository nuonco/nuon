package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type RestartQueueRequest struct{}

// @ID						AdminRestartQueue
// @Summary				restart a queue workflow
// @Description.markdown	restart_queue.md
// @Param					queue_id	path	string					true	"queue ID"
// @Param					req			body	RestartQueueRequest		true	"Input"
// @Tags					queues/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/queues/{queue_id}/admin-restart [POST]
func (s *service) RestartQueue(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")

	var req RestartQueueRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("org id was not found in context"),
			Description: "please make sure you have set your email in the auth login, and that this object is supported in the admin middleware",
		})
		return
	}

	queue, err := s.getQueue(ctx, queueID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get queue: %w", err))
		return
	}

	if err := s.queueClient.Restart(ctx, queue.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to restart queue: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
