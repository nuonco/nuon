package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type SandboxRunsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewSandboxRunsHandler(cfg *internal.Config, l *zap.Logger) *SandboxRunsHandler {
	return &SandboxRunsHandler{cfg: cfg, l: l}
}

func (h *SandboxRunsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/sandbox-runs/:runId/sse", h.StreamSandboxRun)
	return nil
}

func (h *SandboxRunsHandler) StreamSandboxRun(c *gin.Context) {
	installID := c.Param("installId")
	runID := c.Param("runId")

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	var workflowID string

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg:        "failed to fetch sandbox run",
		FinishedGracePeriod: sseFinishedGracePeriod,
		Log:                 h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			sandboxRun, err := client.GetInstallSandboxRun(ctx, installID, runID)
			if err != nil {
				return sseFetchResult{}, err
			}

			primary, err := marshalEvent("sandbox-run", sandboxRun)
			if err != nil {
				return sseFetchResult{}, fmt.Errorf("marshal sandbox run: %v: %w", err, errSSESilentRetry)
			}

			if sandboxRun.InstallWorkflowID != "" {
				workflowID = sandboxRun.InstallWorkflowID
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
			if sandboxRun.StatusV2 != nil {
				status = string(sandboxRun.StatusV2.Status)
			}
			return sseFetchResult{
				Events:   events,
				Finished: terminalStatuses[status],
			}, nil
		},
	})
}
