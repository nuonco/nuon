package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ForceExecuteSignalRequest struct{}

// @ID						AdminForceExecuteQueueSignal
// @Summary				force execute a queue signal
// @Description.markdown	force_execute_queue_signal.md
// @Param					queue_id	path	string						true	"queue ID"
// @Param					signal_id	path	string						true	"signal ID"
// @Param					req			body	ForceExecuteSignalRequest	true	"Input"
// @Tags					queues/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/queues/{queue_id}/signals/{signal_id}/admin-force-execute [POST]
func (s *service) ForceExecuteSignal(ctx *gin.Context) {
	signalID := ctx.Param("signal_id")

	var req ForceExecuteSignalRequest
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

	resp, err := s.queueClient.ForceExecuteSignal(ctx, signalID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to force execute signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
