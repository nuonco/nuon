package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID					GetQueueSignalGraph
// @Summary			Get signal execution graph
// @Description		Returns a recursive tree of a signal and all its awaited/enqueued child signals.
// @Tags				queues
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Param				queue_id	path		string	true	"Queue ID"
// @Param				signal_id	path		string	true	"Signal ID"
// @Param				depth		query		int		false	"Max recursion depth (default 1, max 10)"
// @Success			200			{object}	map[string]interface{}
// @Failure			404			{object}	stderr.ErrResponse
// @Router				/v1/queues/{queue_id}/signals/{signal_id}/graph [get]
func (s *service) GetQueueSignalGraph(ctx *gin.Context) {
	queueID := ctx.Param("queue_id")
	signalID := ctx.Param("signal_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var signal app.QueueSignal
	res := s.db.WithContext(ctx).
		Where(app.QueueSignal{ID: signalID, QueueID: queueID}).
		Where("org_id = ?", org.ID).
		First(&signal)

	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get signal %s: %w", signalID, res.Error))
		return
	}

	maxDepth := 1
	if d := ctx.Query("depth"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 10 {
			maxDepth = n
		}
	}

	node := s.queueHelpers.BuildSignalGraphNode(ctx.Request.Context(), &signal, 0, maxDepth, org.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"graph": node,
	})
}
