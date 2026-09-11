package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type AppBranchRunTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewAppBranchRunTimelineHandler(cfg *internal.Config, l *zap.Logger) *AppBranchRunTimelineHandler {
	return &AppBranchRunTimelineHandler{cfg: cfg, l: l}
}

func (h *AppBranchRunTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/apps/:appId/branches/:branchId/runs/sse", h.StreamAppBranchRunTimeline)
	return nil
}

func appBranchRunsQuery(c *gin.Context) *nuon.GetAppBranchRunsQuery {
	limit, offset := timelineQuery(c)
	planonly := c.DefaultQuery("planonly", "true") == "true"

	return &nuon.GetAppBranchRunsQuery{
		Planonly:     &planonly,
		Type:         c.Query("type"),
		Status:       c.Query("status"),
		Q:            c.Query("q"),
		CreatedAtGte: c.Query("created_at_gte"),
		CreatedAtLte: c.Query("created_at_lte"),
		Limit:        limit,
		Offset:       offset,
	}
}

func (h *AppBranchRunTimelineHandler) StreamAppBranchRunTimeline(c *gin.Context) {
	appID := c.Param("appId")
	branchID := c.Param("branchId")
	query := appBranchRunsQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch branch runs",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: timelineFetcher("branch-runs", func(ctx context.Context) (any, bool, error) {
			return client.GetAppBranchRunsWithQuery(ctx, appID, branchID, query)
		}),
	})
}
