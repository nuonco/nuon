package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type WorkflowTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewWorkflowTimelineHandler(cfg *internal.Config, l *zap.Logger) *WorkflowTimelineHandler {
	return &WorkflowTimelineHandler{cfg: cfg, l: l}
}

func (h *WorkflowTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/workflows/sse", h.StreamWorkflowTimeline)
	return nil
}

func (h *WorkflowTimelineHandler) StreamWorkflowTimeline(c *gin.Context) {
	installID := c.Param("installId")
	limit, offset := timelineQuery(c)
	planonly := c.DefaultQuery("planonly", "true") == "true"
	workflowType := c.DefaultQuery("type", "")
	search := c.DefaultQuery("search", "")

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch workflows",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			var events []sseEvent

			f := false
			if active, _, err := client.GetInstallWorkflows(ctx, installID, &nuon.GetInstallWorkflowsQuery{
				Finished: &f,
				Planonly: &f,
				Limit:    50,
			}); err == nil {
				if ev, err := marshalEvent("active-workflows", active); err == nil {
					events = append(events, ev)
				}
			}

			history, hasMore, err := client.GetInstallWorkflows(ctx, installID, &nuon.GetInstallWorkflowsQuery{
				Planonly: &planonly,
				Type:     workflowType,
				Search:   search,
				Limit:    limit,
				Offset:   offset,
			})
			if err != nil {
				if !isNotFoundErr(err) {
					// Partial result: active-workflows still emits before the
					// fetch-error event.
					return sseFetchResult{Events: events}, err
				}
				history, hasMore = nil, false
			}

			ev, err := marshalEvent("workflows", timelinePayload{
				Data:       history,
				Pagination: paginationInfo{HasNext: hasMore},
			})
			if err != nil {
				return sseFetchResult{Events: events}, fmt.Errorf("marshal workflows: %v: %w", err, errSSESilentRetry)
			}
			return sseFetchResult{Events: append(events, ev)}, nil
		},
	})
}
