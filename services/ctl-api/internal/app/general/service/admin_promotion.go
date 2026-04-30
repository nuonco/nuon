package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/v2/promotion"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type AdminPromotionRequest struct {
	Tag string `json:"tag"`
}

// @ID						AdminPromotion
// @Summary				promotion callback.
// @Description.markdown	promotion.md
// @Param					req	body	AdminPromotionRequest	true	"Input"
// @Tags					general/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/general/promotion [POST]
func (s *service) AdminPromotion(ctx *gin.Context) {
	var req AdminPromotionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	q, err := s.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure general queue: %w", err))
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &promotion.Signal{Tag: req.Tag},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue promotion signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
