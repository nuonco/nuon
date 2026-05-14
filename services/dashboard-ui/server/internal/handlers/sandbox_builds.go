package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type SandboxBuildsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewSandboxBuildsHandler(cfg *internal.Config, l *zap.Logger) *SandboxBuildsHandler {
	return &SandboxBuildsHandler{cfg: cfg, l: l}
}

const (
	sandboxBuildPollInterval         = 2 * time.Second
	sandboxBuildFinishedPollInterval = 30 * time.Second
	sandboxBuildErrorRetryDelay      = 5 * time.Second
	sandboxBuildFinishedGracePeriod  = 2 * time.Minute
)

func (h *SandboxBuildsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/apps/:appId/sandbox-builds/:buildId/sse", h.StreamSandboxBuild)
	return nil
}

type sandboxBuildStatus struct {
	StatusV2 *struct {
		Status string `json:"status"`
	} `json:"status_v2"`
}

func (h *SandboxBuildsHandler) StreamSandboxBuild(c *gin.Context) {
	orgID := c.Param("orgId")
	appID := c.Param("appId")
	buildID := c.Param("buildId")

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ctx := c.Request.Context()
	httpClient := &http.Client{}
	var lastHash string
	var finishedAt time.Time

	sendEvent := func(event string, data string) {
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		c.Writer.Flush()
	}

	fetchBuild := func() ([]byte, error) {
		url := fmt.Sprintf("%s/v1/apps/%s/sandbox/builds/%s", h.cfg.APIUrl, appID, buildID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Nuon-Org-ID", orgID)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("api returned %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		data, err := fetchBuild()
		if err != nil {
			h.l.Error("failed to fetch sandbox build", zap.String("appID", appID), zap.String("buildID", buildID), zap.Error(err))
			sendEvent("fetch-error", `{"error":"failed to fetch sandbox build"}`)
			select {
			case <-ctx.Done():
				return
			case <-time.After(sandboxBuildErrorRetryDelay):
			}
			continue
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

		if hashStr != lastHash {
			lastHash = hashStr
			sendEvent("sandbox-build", string(data))

			var s sandboxBuildStatus
			if json.Unmarshal(data, &s) == nil && s.StatusV2 != nil && deployTerminalStatuses[s.StatusV2.Status] {
				sendEvent("finished", "true")
				if finishedAt.IsZero() {
					finishedAt = time.Now()
				}
			} else if !finishedAt.IsZero() {
				finishedAt = time.Time{}
			}
		}

		if !finishedAt.IsZero() && time.Since(finishedAt) > sandboxBuildFinishedGracePeriod {
			return
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		interval := sandboxBuildPollInterval
		if !finishedAt.IsZero() {
			interval = sandboxBuildFinishedPollInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
