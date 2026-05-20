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
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const (
	buildTimelinePollInterval    = 3 * time.Second
	buildTimelineErrorRetryDelay = 5 * time.Second
)

type BuildTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewBuildTimelineHandler(cfg *internal.Config, l *zap.Logger) *BuildTimelineHandler {
	return &BuildTimelineHandler{cfg: cfg, l: l}
}

func (h *BuildTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/components/:componentId/builds/sse", h.StreamBuildTimeline)
	return nil
}

func (h *BuildTimelineHandler) StreamBuildTimeline(c *gin.Context) {
	orgID := c.Param("orgId")
	componentID := c.Param("componentId")
	appID := c.Query("appId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

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

		builds, hasMore, err := client.GetComponentBuilds(ctx, componentID, appID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
		if err != nil {
			if isNotFoundErr(err) {
				builds = nil
				hasMore = false
			} else {
				h.l.Error("failed to fetch builds", zap.String("componentID", componentID), zap.Error(err))
				sendEvent("fetch-error", `{"error":"failed to fetch builds"}`)
				select {
				case <-ctx.Done():
					return
				case <-time.After(buildTimelineErrorRetryDelay):
				}
				continue
			}
		}

		payload := timelinePayload{Data: builds, Pagination: paginationInfo{HasNext: hasMore}}
		data, err := json.Marshal(payload)
		if err != nil {
			h.l.Error("failed to marshal builds", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(buildTimelineErrorRetryDelay):
			}
			continue
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

		if hashStr != lastHash {
			lastHash = hashStr
			sendEvent("builds", string(data))
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		select {
		case <-ctx.Done():
			return
		case <-time.After(buildTimelinePollInterval):
		}
	}
}
