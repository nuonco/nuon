package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type DeploysHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewDeploysHandler(cfg *internal.Config, l *zap.Logger) *DeploysHandler {
	return &DeploysHandler{cfg: cfg, l: l}
}

func (h *DeploysHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/deploys/:deployId/sse", h.StreamDeploy)
	return nil
}

func (h *DeploysHandler) StreamDeploy(c *gin.Context) {
	installID := c.Param("installId")
	deployID := c.Param("deployId")

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	var componentID, workflowID string

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg:        "failed to fetch deploy",
		FinishedGracePeriod: sseFinishedGracePeriod,
		Log:                 h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			deploy, err := client.GetInstallDeploy(ctx, installID, deployID)
			if err != nil {
				return sseFetchResult{}, err
			}

			primary, err := marshalEvent("deploy", deploy)
			if err != nil {
				return sseFetchResult{}, fmt.Errorf("marshal deploy: %v: %w", err, errSSESilentRetry)
			}

			if deploy.ComponentID != "" {
				componentID = deploy.ComponentID
			}
			if deploy.InstallWorkflowID != "" {
				workflowID = deploy.InstallWorkflowID
			}

			events := []sseEvent{primary}
			if componentID != "" {
				if component, err := client.GetComponent(ctx, componentID); err == nil {
					if ev, err := marshalEvent("component", component); err == nil {
						events = append(events, ev)
					}
				}
			}
			if workflowID != "" {
				if workflow, err := client.GetWorkflow(ctx, workflowID); err == nil {
					if ev, err := marshalEvent("workflow", workflow); err == nil {
						events = append(events, ev)
					}
				}
			}

			status := ""
			if deploy.StatusV2 != nil {
				status = string(deploy.StatusV2.Status)
			}
			return sseFetchResult{
				Events:   events,
				Finished: terminalStatuses[status],
			}, nil
		},
	})
}
