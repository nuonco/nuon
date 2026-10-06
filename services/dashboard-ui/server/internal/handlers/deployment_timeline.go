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

const deploymentsPageLimit = 100

func deploymentTimelineQuery(c *gin.Context) *nuon.GetInstallDeploymentsQuery {
	limit, offset := timelineQuery(c)
	limit = max(limit, 1)
	return &nuon.GetInstallDeploymentsQuery{
		Type:         c.Query("type"),
		Status:       c.Query("status"),
		Resource:     c.Query("resource"),
		Search:       c.Query("search"),
		CreatedAtGte: c.Query("created_at_gte"),
		CreatedAtLte: c.Query("created_at_lte"),
		State:        c.Query("state"),
		Sort:         c.Query("sort"),
		Cursor:       c.Query("cursor"),
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
			deployments, err := fetchDeploymentWindow(ctx, client, installID, query)
			if err != nil {
				if !isNotFoundErr(err) {
					return sseFetchResult{}, err
				}
				deployments = &models.ServiceGetInstallDeploymentsResponse{
					Deployments: []*models.ServiceInstallDeploymentSummary{},
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

func fetchDeploymentWindow(ctx context.Context, client nuon.Client, installID string, query *nuon.GetInstallDeploymentsQuery) (*models.ServiceGetInstallDeploymentsResponse, error) {
	chunk := *query
	chunk.Limit = min(query.Limit, deploymentsPageLimit)
	window, err := client.GetInstallDeployments(ctx, installID, &chunk)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(window.Deployments))
	for _, deployment := range window.Deployments {
		seen[deployment.ID] = struct{}{}
	}
	for remaining := query.Limit - len(window.Deployments); remaining > 0 && window.HasMore && window.NextCursor != ""; remaining = query.Limit - len(window.Deployments) {
		chunk.Cursor = window.NextCursor
		chunk.Offset = 0
		chunk.Limit = min(remaining, deploymentsPageLimit)
		next, err := client.GetInstallDeployments(ctx, installID, &chunk)
		if err != nil {
			return nil, err
		}
		for _, deployment := range next.Deployments {
			if _, ok := seen[deployment.ID]; !ok {
				seen[deployment.ID] = struct{}{}
				window.Deployments = append(window.Deployments, deployment)
			}
		}
		window.HasMore = next.HasMore
		window.NextCursor = next.NextCursor
	}

	window.Limit = int64(query.Limit)
	return window, nil
}
