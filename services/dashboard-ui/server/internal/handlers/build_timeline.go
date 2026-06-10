package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type BuildTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewBuildTimelineHandler(cfg *internal.Config, l *zap.Logger) *BuildTimelineHandler {
	return &BuildTimelineHandler{cfg: cfg, l: l}
}

func (h *BuildTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/components/:componentId/builds/sse", h.StreamBuildTimeline)
	return nil
}

func (h *BuildTimelineHandler) StreamBuildTimeline(c *gin.Context) {
	componentID := c.Param("componentId")
	appID := c.Query("appId")
	limit, offset := timelineQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch builds",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: timelineFetcher("builds", func(ctx context.Context) (any, bool, error) {
			builds, hasMore, err := client.GetComponentBuilds(ctx, componentID, appID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
			return builds, hasMore, err
		}),
	})
}
