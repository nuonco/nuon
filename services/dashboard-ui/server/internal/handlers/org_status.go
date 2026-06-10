package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type OrgStatusHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewOrgStatusHandler(cfg *internal.Config, l *zap.Logger) *OrgStatusHandler {
	return &OrgStatusHandler{cfg: cfg, l: l}
}

func (h *OrgStatusHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/status/sse", h.StreamOrgStatus)
	return nil
}

func (h *OrgStatusHandler) StreamOrgStatus(c *gin.Context) {
	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	var runnerID string

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch org",
		PollInterval: sseOrgStatusPollInterval,
		Log:          h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			org, err := client.GetOrg(ctx)
			if err != nil {
				return sseFetchResult{}, err
			}

			var events []sseEvent
			if ev, err := marshalEvent("org", org); err == nil {
				events = append(events, ev)
			}

			if runnerID == "" && org.RunnerGroup != nil {
				for _, r := range org.RunnerGroup.Runners {
					if r != nil && r.ID != "" {
						runnerID = r.ID
						break
					}
				}
			}

			if workflows, err := client.GetOrgWorkflows(ctx, &nuon.GetOrgWorkflowsQuery{
				Finished: false,
				Planonly: false,
				Limit:    50,
			}); err == nil {
				if ev, err := marshalEvent("active-workflows", workflows); err == nil {
					events = append(events, ev)
				}
			}

			if approvals, err := client.GetOrgPendingApprovals(ctx); err == nil {
				if ev, err := marshalEvent("pending-approvals", approvals); err == nil {
					events = append(events, ev)
				}
			}

			if runnerID != "" {
				if heartbeats, err := client.GetLatestRunnerHeartBeats(ctx, runnerID); err == nil {
					if ev, err := marshalEvent("runner-heartbeat", heartbeats); err == nil {
						events = append(events, ev)
					}
				}
			}

			return sseFetchResult{Events: events}, nil
		},
	})
}
