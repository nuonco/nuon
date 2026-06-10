package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type ActionRunTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewActionRunTimelineHandler(cfg *internal.Config, l *zap.Logger) *ActionRunTimelineHandler {
	return &ActionRunTimelineHandler{cfg: cfg, l: l}
}

func (h *ActionRunTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/actions/:actionId/runs/sse", h.StreamActionRunTimeline)
	return nil
}

func (h *ActionRunTimelineHandler) StreamActionRunTimeline(c *gin.Context) {
	installID := c.Param("installId")
	actionID := c.Param("actionId")
	limit, offset := timelineQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch action runs",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: timelineFetcher("action-runs", func(ctx context.Context) (any, bool, error) {
			runs, hasMore, err := client.GetInstallActionWorkflowRecentRuns(ctx, installID, actionID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
			return runs, hasMore, err
		}),
	})
}
