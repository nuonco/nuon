package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID				GetQueue
// @Summary		Get queue by ID
// @Description	Retrieve a single queue by its ID
// @Tags			queues
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			queue_id	path		string	true	"Queue ID"
// @Success		200			{object}	app.Queue
// @Failure		404			{object}	stderr.ErrResponse
// @Router			/v1/queues/{queue_id} [get]
func (s *service) GetQueue(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	queue, err := s.queueClient.GetQueue(ctx, queueID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Verify queue belongs to the org
	if queue.OrgID == nil || *queue.OrgID != org.ID {
		ctx.Error(stderr.ErrNotFound{
			Err:         errors.New("queue does not belong to organization"),
			Description: "Queue not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, queue)
}
