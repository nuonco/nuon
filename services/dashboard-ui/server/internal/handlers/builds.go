package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type BuildsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewBuildsHandler(cfg *internal.Config, l *zap.Logger) *BuildsHandler {
	return &BuildsHandler{cfg: cfg, l: l}
}

func (h *BuildsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/components/:componentId/builds/:buildId/sse", h.StreamBuild)
	return nil
}

func (h *BuildsHandler) StreamBuild(c *gin.Context) {
	componentID := c.Param("componentId")
	buildID := c.Param("buildId")

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg:        "failed to fetch build",
		FinishedGracePeriod: sseFinishedGracePeriod,
		Log:                 h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			build, err := client.GetComponentBuild(ctx, componentID, buildID)
			if err != nil {
				return sseFetchResult{}, err
			}

			ev, err := marshalEvent("build", build)
			if err != nil {
				return sseFetchResult{}, fmt.Errorf("marshal build: %v: %w", err, errSSESilentRetry)
			}

			status := ""
			if build.StatusV2 != nil {
				status = string(build.StatusV2.Status)
			}
			return sseFetchResult{
				Events:   []sseEvent{ev},
				Finished: terminalStatuses[status],
			}, nil
		},
	})
}
