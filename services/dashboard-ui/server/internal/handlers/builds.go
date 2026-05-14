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

type BuildsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewBuildsHandler(cfg *internal.Config, l *zap.Logger) *BuildsHandler {
	return &BuildsHandler{cfg: cfg, l: l}
}

const (
	buildPollInterval         = 2 * time.Second
	buildFinishedPollInterval = 30 * time.Second
	buildErrorRetryDelay      = 5 * time.Second
	buildFinishedGracePeriod  = 2 * time.Minute
)

func (h *BuildsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/components/:componentId/builds/:buildId/sse", h.StreamBuild)
	return nil
}

func (h *BuildsHandler) StreamBuild(c *gin.Context) {
	orgID := c.Param("orgId")
	componentID := c.Param("componentId")
	buildID := c.Param("buildId")

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

		build, err := client.GetComponentBuild(ctx, componentID, buildID)
		if err != nil {
			h.l.Error("failed to fetch build", zap.String("componentID", componentID), zap.String("buildID", buildID), zap.Error(err))
			sendEvent("fetch-error", `{"error":"failed to fetch build"}`)
			select {
			case <-ctx.Done():
				return
			case <-time.After(buildErrorRetryDelay):
			}
			continue
		}

		data, err := json.Marshal(build)
		if err != nil {
			h.l.Error("failed to marshal build", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(buildErrorRetryDelay):
			}
			continue
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

		if hashStr != lastHash {
			lastHash = hashStr
			sendEvent("build", string(data))

			status := ""
			if build.StatusV2 != nil {
				status = string(build.StatusV2.Status)
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

		if !finishedAt.IsZero() && time.Since(finishedAt) > buildFinishedGracePeriod {
			return
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		interval := buildPollInterval
		if !finishedAt.IsZero() {
			interval = buildFinishedPollInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
