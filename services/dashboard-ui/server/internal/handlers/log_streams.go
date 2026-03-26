package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const maxDownloadLogs = 50000

const (
	streamingThreshold = 40
	streamingDelay     = 200 * time.Millisecond
	pollInterval       = 1 * time.Second
	errorRetryDelay    = 5 * time.Second
)

type LogStreamsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewLogStreamsHandler(cfg *internal.Config, l *zap.Logger) *LogStreamsHandler {
	return &LogStreamsHandler{cfg: cfg, l: l}
}

func (h *LogStreamsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/log-streams/:logStreamId/logs/sse", h.StreamLogs)
	e.GET("/api/orgs/:orgId/log-streams/:logStreamId/logs/download", h.DownloadLogs)
	return nil
}

func (h *LogStreamsHandler) StreamLogs(c *gin.Context) {
	orgID := c.Param("orgId")
	logStreamID := c.Param("logStreamId")

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
	var currentOffset string
	isCatchingUp := false
	hasSeenFirstBatch := false

	sendEvent := func(logs []*models.AppOtelLogRecord) {
		b, _ := json.Marshal(logs)
		fmt.Fprintf(c.Writer, "data: %s\n\n", b)
		c.Writer.Flush()
	}

	sendStatus := func(status string) {
		fmt.Fprintf(c.Writer, "event: status\ndata: %s\n\n", status)
		c.Writer.Flush()
	}

	sendError := func(msg string) {
		b, _ := json.Marshal(map[string]string{"error": msg})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", b)
		c.Writer.Flush()
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		logs, nextOffset, err := client.LogStreamReadLogsWithNextOffset(ctx, logStreamID, currentOffset)
		if err != nil {
			sendError("Polling failed")
			select {
			case <-ctx.Done():
				return
			case <-time.After(errorRetryDelay):
			}
			continue
		}

		if nextOffset != "" {
			currentOffset = nextOffset
		}

		if len(logs) > 0 {
			if !hasSeenFirstBatch {
				isCatchingUp = len(logs) >= streamingThreshold
				hasSeenFirstBatch = true
				if isCatchingUp {
					sendStatus("catching-up")
				}
			}

			paginationComplete := nextOffset == ""

			if isCatchingUp {
				sendEvent(logs)
				if paginationComplete {
					isCatchingUp = false
					sendStatus("live")
				} else {
					continue
				}
			} else {
				for _, log := range logs {
					select {
					case <-ctx.Done():
						return
					case <-time.After(streamingDelay):
					}
					sendEvent([]*models.AppOtelLogRecord{log})
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(pollInterval):
		}
	}
}

func (h *LogStreamsHandler) DownloadLogs(c *gin.Context) {
	orgID := c.Param("orgId")
	logStreamID := c.Param("logStreamId")

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

	ctx := c.Request.Context()
	var offset string
	totalLogs := 0

	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="logs-%s.txt"`, logStreamID))
	c.Writer.WriteHeader(http.StatusOK)

	jobOutputOnly := c.Query("job_output") == "true"

	for {
		logs, nextOffset, err := client.LogStreamReadLogsWithNextOffset(ctx, logStreamID, offset)
		if err != nil {
			h.l.Error("failed to read log stream logs", zap.Error(err), zap.String("logStreamID", logStreamID))
			break
		}

		for _, log := range logs {
			if jobOutputOnly && log.ScopeName != "oteljob" {
				continue
			}
			fmt.Fprintf(c.Writer, "[%s] [%s] [%s] %s\n", log.Timestamp, log.SeverityText, log.ServiceName, log.Body)
		}
		c.Writer.Flush()

		totalLogs += len(logs)
		if nextOffset == "" || totalLogs >= maxDownloadLogs {
			break
		}
		offset = nextOffset
	}
}
