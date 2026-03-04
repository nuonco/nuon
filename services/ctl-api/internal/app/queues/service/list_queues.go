package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ListQueuesRequest struct {
	OwnerID   string `form:"owner_id"`
	OwnerType string `form:"owner_type"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// @ID				ListQueues
// @Summary		List queues
// @Description	List queues with optional filtering by owner
// @Tags			queues
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			owner_id	query		string	false	"Filter by owner ID"
// @Param			owner_type	query		string	false	"Filter by owner type (e.g., 'app_branches')"
// @Param			limit		query		int		false	"Limit results"	default(50)
// @Param			offset		query		int		false	"Offset results"	default(0)
// @Success		200			{array}		app.Queue
// @Failure		400			{object}	stderr.ErrResponse
// @Router			/v1/queues [get]
func (s *service) ListQueues(ctx *gin.Context) {
	var req ListQueuesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.Error(err)
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 50
	}

	queues, err := s.queueClient.ListQueues(ctx, org.ID, req.OwnerID, req.OwnerType, req.Limit, req.Offset)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, queues)
}
