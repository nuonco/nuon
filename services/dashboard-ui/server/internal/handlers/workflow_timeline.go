package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const (
	workflowTimelinePollInterval    = 3 * time.Second
	workflowTimelineErrorRetryDelay = 5 * time.Second
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
	orgID := c.Param("orgId")
	installID := c.Param("installId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	planonlyStr := c.DefaultQuery("planonly", "true")
	planonly := planonlyStr == "true"
	workflowType := c.DefaultQuery("type", "")

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
	var activeHash, historyHash string

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

		finished := false
		activeLimit := int64(50)
		activeWorkflows, err := client.GetOrgWorkflows(ctx, &nuon.GetOrgWorkflowsQuery{
			Finished: finished,
			Planonly: false,
			Limit:    activeLimit,
		})
		if err == nil {
			aData, aHash, err := hashJSON(activeWorkflows)
			if err == nil && aHash != activeHash {
				activeHash = aHash
				sendEvent("active-workflows", string(aData))
			}
		}

		historyWorkflows, hasMore, err := client.GetInstallWorkflows(ctx, installID, &nuon.GetInstallWorkflowsQuery{
			Planonly: &planonly,
			Type:     workflowType,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			if isNotFoundErr(err) {
				historyWorkflows = nil
				hasMore = false
			} else {
				h.l.Error("failed to fetch workflow history", zap.String("installID", installID), zap.Error(err))
				sendEvent("fetch-error", `{"error":"failed to fetch workflows"}`)
				select {
				case <-ctx.Done():
					return
				case <-time.After(workflowTimelineErrorRetryDelay):
				}
				continue
			}
		}

		payload := timelinePayload{Data: historyWorkflows, Pagination: paginationInfo{HasNext: hasMore}}
		hData, hHash, err := hashJSON(payload)
		if err == nil && hHash != historyHash {
			historyHash = hHash
			sendEvent("workflows", string(hData))
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		select {
		case <-ctx.Done():
			return
		case <-time.After(workflowTimelinePollInterval):
		}
	}
}
