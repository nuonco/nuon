package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type SandboxRunTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewSandboxRunTimelineHandler(cfg *internal.Config, l *zap.Logger) *SandboxRunTimelineHandler {
	return &SandboxRunTimelineHandler{cfg: cfg, l: l}
}

func (h *SandboxRunTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/sandbox-runs/sse", h.StreamSandboxRunTimeline)
	return nil
}

func (h *SandboxRunTimelineHandler) StreamSandboxRunTimeline(c *gin.Context) {
	installID := c.Param("installId")
	limit, offset := timelineQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch sandbox runs",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: timelineFetcher("sandbox-runs", func(ctx context.Context) (any, bool, error) {
			runs, hasMore, err := client.GetInstallSandboxRuns(ctx, installID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
			return runs, hasMore, err
		}),
	})
}
