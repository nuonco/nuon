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

type SandboxRunsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewSandboxRunsHandler(cfg *internal.Config, l *zap.Logger) *SandboxRunsHandler {
	return &SandboxRunsHandler{cfg: cfg, l: l}
}

const (
	sandboxRunPollInterval         = 2 * time.Second
	sandboxRunFinishedPollInterval = 30 * time.Second
	sandboxRunErrorRetryDelay      = 5 * time.Second
	sandboxRunFinishedGracePeriod  = 2 * time.Minute
)

func (h *SandboxRunsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/sandbox-runs/:runId/sse", h.StreamSandboxRun)
	return nil
}

func (h *SandboxRunsHandler) StreamSandboxRun(c *gin.Context) {
	orgID := c.Param("orgId")
	installID := c.Param("installId")
	runID := c.Param("runId")

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
	var sandboxRunHash, workflowHash string
	var workflowID string
	var finishedAt time.Time

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

		sandboxRun, err := client.GetInstallSandboxRun(ctx, installID, runID)
		if err != nil {
			h.l.Error("failed to fetch sandbox run", zap.String("installID", installID), zap.String("runID", runID), zap.Error(err))
			sendEvent("fetch-error", `{"error":"failed to fetch sandbox run"}`)
			select {
			case <-ctx.Done():
				return
			case <-time.After(sandboxRunErrorRetryDelay):
			}
			continue
		}

		runData, rHash, err := hashJSON(sandboxRun)
		if err != nil {
			h.l.Error("failed to marshal sandbox run", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(sandboxRunErrorRetryDelay):
			}
			continue
		}

		if rHash != sandboxRunHash {
			sandboxRunHash = rHash
			sendEvent("sandbox-run", string(runData))

			if sandboxRun.InstallWorkflowID != "" {
				workflowID = sandboxRun.InstallWorkflowID
			}

			status := ""
			if sandboxRun.StatusV2 != nil {
				status = string(sandboxRun.StatusV2.Status)
			}
			if deployTerminalStatuses[status] {
				sendEvent("finished", "true")
				if finishedAt.IsZero() {
					finishedAt = time.Now()
				}
			} else if !finishedAt.IsZero() {
				finishedAt = time.Time{}
			}
		}

		if workflowID != "" {
			workflow, err := client.GetWorkflow(ctx, workflowID)
			if err == nil {
				wData, wHash, err := hashJSON(workflow)
				if err == nil && wHash != workflowHash {
					workflowHash = wHash
					sendEvent("workflow", string(wData))
				}
			}
		}

		if !finishedAt.IsZero() && time.Since(finishedAt) > sandboxRunFinishedGracePeriod {
			return
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		interval := sandboxRunPollInterval
		if !finishedAt.IsZero() {
			interval = sandboxRunFinishedPollInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
