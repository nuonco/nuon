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

type WorkflowsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewWorkflowsHandler(cfg *internal.Config, l *zap.Logger) *WorkflowsHandler {
	return &WorkflowsHandler{cfg: cfg, l: l}
}

const (
	workflowPollInterval         = 2 * time.Second
	workflowFinishedPollInterval = 30 * time.Second
	workflowErrorRetryDelay      = 5 * time.Second
	workflowFinishedGracePeriod  = 2 * time.Minute
)

func (h *WorkflowsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/workflows/:workflowId/steps/:stepId/approvals/:approvalId/contents", h.GetApprovalContents)
	e.GET("/api/orgs/:orgId/workflows/:workflowId/sse", h.StreamWorkflow)
	return nil
}

func (h *WorkflowsHandler) StreamWorkflow(c *gin.Context) {
	orgID := c.Param("orgId")
	workflowID := c.Param("workflowId")

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
	var lastHash string
	var finishedAt time.Time

	sendData := func(data []byte) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.Flush()
	}

	sendEvent := func(event string, data string) {
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		c.Writer.Flush()
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		workflow, err := client.GetWorkflow(ctx, workflowID)
		if err != nil {
			sendEvent("fetch-error", `{"error":"failed to fetch workflow"}`)
			select {
			case <-ctx.Done():
				return
			case <-time.After(workflowErrorRetryDelay):
			}
			continue
		}

		data, err := json.Marshal(workflow)
		if err != nil {
			h.l.Error("failed to marshal workflow", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(workflowErrorRetryDelay):
			}
			continue
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

		if hashStr != lastHash {
			lastHash = hashStr
			sendData(data)

			if workflow.Finished {
				sendEvent("finished", "true")
				finishedAt = time.Now()
			} else if !finishedAt.IsZero() {
				finishedAt = time.Time{}
			}
		}

		if !finishedAt.IsZero() && time.Since(finishedAt) > workflowFinishedGracePeriod {
			return
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		interval := workflowPollInterval
		if !finishedAt.IsZero() {
			interval = workflowFinishedPollInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

func (h *WorkflowsHandler) GetApprovalContents(c *gin.Context) {
	orgID := c.Param("orgId")
	workflowID := c.Param("workflowId")
	stepID := c.Param("stepId")
	approvalID := c.Param("approvalId")

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

	contents, err := client.GetWorkflowStepApprovalContents(c.Request.Context(), workflowID, stepID, approvalID)
	if err != nil {
		h.l.Error("failed to get approval contents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contents)
}
