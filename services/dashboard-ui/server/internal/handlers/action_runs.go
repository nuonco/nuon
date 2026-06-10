package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type ActionRunsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewActionRunsHandler(cfg *internal.Config, l *zap.Logger) *ActionRunsHandler {
	return &ActionRunsHandler{cfg: cfg, l: l}
}

func (h *ActionRunsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/action-runs/:runId/sse", h.StreamActionRun)
	return nil
}

func (h *ActionRunsHandler) StreamActionRun(c *gin.Context) {
	installID := c.Param("installId")
	runID := c.Param("runId")

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	var workflowID string

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg:        "failed to fetch action run",
		FinishedGracePeriod: sseFinishedGracePeriod,
		Log:                 h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			actionRun, err := client.GetInstallActionWorkflowRun(ctx, installID, runID)
			if err != nil {
				return sseFetchResult{}, err
			}

			primary, err := marshalEvent("action-run", actionRun)
			if err != nil {
				return sseFetchResult{}, fmt.Errorf("marshal action run: %v: %w", err, errSSESilentRetry)
			}

			if actionRun.InstallWorkflowID != "" {
				workflowID = actionRun.InstallWorkflowID
			}

			events := []sseEvent{primary}
			if workflowID != "" {
				if workflow, err := client.GetWorkflow(ctx, workflowID); err == nil {
					if ev, err := marshalEvent("workflow", workflow); err == nil {
						events = append(events, ev)
					}
				}
			}

			status := ""
			if actionRun.StatusV2 != nil {
				status = string(actionRun.StatusV2.Status)
			}
			return sseFetchResult{
				Events:   events,
				Finished: terminalStatuses[status],
			}, nil
		},
	})
}
