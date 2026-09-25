package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type DeploymentTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewDeploymentTimelineHandler(cfg *internal.Config, l *zap.Logger) *DeploymentTimelineHandler {
	return &DeploymentTimelineHandler{cfg: cfg, l: l}
}

func (h *DeploymentTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/deployments/sse", h.StreamDeploymentTimeline)
	return nil
}

func deploymentTimelineQuery(c *gin.Context) *nuon.GetInstallDeploymentsQuery {
	limit, offset := timelineQuery(c)
	return &nuon.GetInstallDeploymentsQuery{
		Type:         c.Query("type"),
		Status:       c.Query("status"),
		Resource:     c.Query("resource"),
		Search:       c.Query("search"),
		CreatedAtGte: c.Query("created_at_gte"),
		CreatedAtLte: c.Query("created_at_lte"),
		Limit:        limit,
		Offset:       offset,
	}
}

func (h *DeploymentTimelineHandler) StreamDeploymentTimeline(c *gin.Context) {
	installID := c.Param("installId")
	query := deploymentTimelineQuery(c)

	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch deployments",
		PollInterval: sseTimelinePollInterval,
		Log:          h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			deployments, err := client.GetInstallDeployments(ctx, installID, query)
			if err != nil {
				if !isNotFoundErr(err) {
					return sseFetchResult{}, err
				}
				deployments = &models.ServiceGetInstallDeploymentsResponse{
					Deployments: []*models.ServiceInstallDeployment{},
					Limit:       int64(query.Limit),
					Offset:      int64(query.Offset),
				}
			}

			ev, err := marshalEvent("deployments", deployments)
			if err != nil {
				return sseFetchResult{}, fmt.Errorf("marshal deployments: %v: %w", err, errSSESilentRetry)
			}
			return sseFetchResult{Events: []sseEvent{ev}}, nil
		},
	})
}
