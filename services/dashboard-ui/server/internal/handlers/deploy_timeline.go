package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type DeployTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewDeployTimelineHandler(cfg *internal.Config, l *zap.Logger) *DeployTimelineHandler {
	return &DeployTimelineHandler{cfg: cfg, l: l}
}

func (h *DeployTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/components/:componentId/deploys/sse", h.StreamDeployTimeline)
	return nil
}

func (h *DeployTimelineHandler) StreamDeployTimeline(c *gin.Context) {
	installID := c.Param("installId")
	componentID := c.Param("componentId")
	limit, offset := timelineQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch deploys",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: timelineFetcher("deploys", func(ctx context.Context) (any, bool, error) {
			deploys, hasMore, err := client.GetInstallComponentDeploys(ctx, installID, componentID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
			return deploys, hasMore, err
		}),
	})
}
