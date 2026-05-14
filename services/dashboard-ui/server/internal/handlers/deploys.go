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

type DeploysHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewDeploysHandler(cfg *internal.Config, l *zap.Logger) *DeploysHandler {
	return &DeploysHandler{cfg: cfg, l: l}
}

const (
	deployPollInterval         = 2 * time.Second
	deployFinishedPollInterval = 30 * time.Second
	deployErrorRetryDelay      = 5 * time.Second
	deployFinishedGracePeriod  = 2 * time.Minute
)

var deployTerminalStatuses = map[string]bool{
	"success":       true,
	"error":         true,
	"failed":        true,
	"cancelled":     true,
	"not-attempted": true,
}

func (h *DeploysHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/deploys/:deployId/sse", h.StreamDeploy)
	return nil
}

func (h *DeploysHandler) StreamDeploy(c *gin.Context) {
	orgID := c.Param("orgId")
	installID := c.Param("installId")
	deployID := c.Param("deployId")

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
	var deployHash, componentHash, workflowHash string
	var componentID, workflowID string
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

		deploy, err := client.GetInstallDeploy(ctx, installID, deployID)
		if err != nil {
			sendEvent("fetch-error", `{"error":"failed to fetch deploy"}`)
			select {
			case <-ctx.Done():
				return
			case <-time.After(deployErrorRetryDelay):
			}
			continue
		}

		deployData, dHash, err := hashJSON(deploy)
		if err != nil {
			h.l.Error("failed to marshal deploy", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(deployErrorRetryDelay):
			}
			continue
		}

		if dHash != deployHash {
			deployHash = dHash
			sendEvent("deploy", string(deployData))

			if deploy.ComponentID != "" {
				componentID = deploy.ComponentID
			}
			if deploy.InstallWorkflowID != "" {
				workflowID = deploy.InstallWorkflowID
			}

			status := ""
			if deploy.StatusV2 != nil {
				status = string(deploy.StatusV2.Status)
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

		if componentID != "" {
			component, err := client.GetComponent(ctx, componentID)
			if err == nil {
				cData, cHash, err := hashJSON(component)
				if err == nil && cHash != componentHash {
					componentHash = cHash
					sendEvent("component", string(cData))
				}
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

		if !finishedAt.IsZero() && time.Since(finishedAt) > deployFinishedGracePeriod {
			return
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		interval := deployPollInterval
		if !finishedAt.IsZero() {
			interval = deployFinishedPollInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
