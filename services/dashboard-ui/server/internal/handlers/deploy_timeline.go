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
	deployTimelinePollInterval    = 3 * time.Second
	deployTimelineErrorRetryDelay = 5 * time.Second
)

type DeployTimelineHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewDeployTimelineHandler(cfg *internal.Config, l *zap.Logger) *DeployTimelineHandler {
	return &DeployTimelineHandler{cfg: cfg, l: l}
}

func (h *DeployTimelineHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs/:installId/components/:componentId/deploys/sse", h.StreamDeployTimeline)
	return nil
}

func (h *DeployTimelineHandler) StreamDeployTimeline(c *gin.Context) {
	orgID := c.Param("orgId")
	installID := c.Param("installId")
	componentID := c.Param("componentId")
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

		deploys, hasMore, err := client.GetInstallComponentDeploys(ctx, installID, componentID, &models.GetPaginatedQuery{Limit: limit, Offset: offset})
		if err != nil {
			if isNotFoundErr(err) {
				deploys = nil
				hasMore = false
			} else {
				h.l.Error("failed to fetch deploys", zap.String("installID", installID), zap.String("componentID", componentID), zap.Error(err))
				sendEvent("fetch-error", `{"error":"failed to fetch deploys"}`)
				select {
				case <-ctx.Done():
					return
				case <-time.After(deployTimelineErrorRetryDelay):
				}
				continue
			}
		}

		payload := timelinePayload{Data: deploys, Pagination: paginationInfo{HasNext: hasMore}}
		data, err := json.Marshal(payload)
		if err != nil {
			h.l.Error("failed to marshal deploys", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(deployTimelineErrorRetryDelay):
			}
			continue
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

		if hashStr != lastHash {
			lastHash = hashStr
			sendEvent("deploys", string(data))
		}

		fmt.Fprintf(c.Writer, ": keepalive\n\n")
		c.Writer.Flush()

		select {
		case <-ctx.Done():
			return
		case <-time.After(deployTimelinePollInterval):
		}
	}
}
