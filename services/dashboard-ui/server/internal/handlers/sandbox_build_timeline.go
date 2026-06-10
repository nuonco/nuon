package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type SandboxBuildTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewSandboxBuildTimelineHandler(cfg *internal.Config, l *zap.Logger) *SandboxBuildTimelineHandler {
	return &SandboxBuildTimelineHandler{cfg: cfg, l: l}
}

func (h *SandboxBuildTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/apps/:appId/sandbox-builds/sse", h.StreamSandboxBuildTimeline)
	return nil
}

func (h *SandboxBuildTimelineHandler) StreamSandboxBuildTimeline(c *gin.Context) {
	appID := c.Param("appId")
	limit, offset := timelineQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch sandbox builds",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: timelineFetcher("sandbox-builds", func(ctx context.Context) (any, bool, error) {
			builds, hasMore, err := client.GetAppSandboxBuilds(ctx, appID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
			return builds, hasMore, err
		}),
	})
}
