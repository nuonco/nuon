package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

const (
	orgStatusPollInterval    = 5 * time.Second
	orgStatusErrorRetryDelay = 5 * time.Second
)

func (h *OrgStatusHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/status/sse", h.StreamOrgStatus)
	return nil
}

func (h *OrgStatusHandler) StreamOrgStatus(c *gin.Context) {
	orgID := c.Param("orgId")

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	client, err := nuon.New(
		nuon.WithURL(h.cfg.APIUrl),
		nuon.WithAuthToken(token),
		nuon.WithOrgID(orgID),
	)
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ctx := c.Request.Context()
	var orgHash, workflowsHash, approvalsHash, heartbeatHash string
	var runnerID string

	sendEvent := func(event string, data string) {
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		c.Writer.Flush()
	}

	hashJSON := func(v any) ([]byte, string, error) {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, "", err
		}
		h := sha256.Sum256(data)
		return data, hex.EncodeToString(h[:]), nil
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		org, err := client.GetOrg(ctx)
		if err != nil {
			sendEvent("fetch-error", `{"error":"failed to fetch org"}`)
			select {
			case <-ctx.Done():
				return
			case <-time.After(orgStatusErrorRetryDelay):
			}
			continue
		}

		orgData, oHash, err := hashJSON(org)
		if err == nil && oHash != orgHash {
			orgHash = oHash
			sendEvent("org", string(orgData))

			if runnerID == "" && org.RunnerGroup != nil {
				for _, r := range org.RunnerGroup.Runners {
					if r != nil && r.ID != "" {
						runnerID = r.ID
						break
					}
				}
			}
		}

		finished := false
		limit := int64(50)
		workflows, err := client.GetOrgWorkflows(ctx, &nuon.GetOrgWorkflowsQuery{
			Finished: finished,
			Planonly: false,
			Limit:    limit,
		})
		if err == nil {
			wData, wHash, err := hashJSON(workflows)
			if err == nil && wHash != workflowsHash {
				workflowsHash = wHash
				sendEvent("active-workflows", string(wData))
			}
		}

		approvals, err := client.GetOrgPendingApprovals(ctx)
		if err == nil {
			aData, aHash, err := hashJSON(approvals)
			if err == nil && aHash != approvalsHash {
				approvalsHash = aHash
				sendEvent("pending-approvals", string(aData))
			}
		}

		if runnerID != "" {
			heartbeats, err := client.GetLatestRunnerHeartBeats(ctx, runnerID)
			if err == nil {
				hData, hHash, err := hashJSON(heartbeats)
				if err == nil && hHash != heartbeatHash {
					heartbeatHash = hHash
					sendEvent("runner-heartbeat", string(hData))
				}
			}
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		select {
		case <-ctx.Done():
			return
		case <-time.After(orgStatusPollInterval):
		}
	}
}
