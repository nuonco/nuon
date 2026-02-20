package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

const (
	streamingThreshold = 40
	streamingDelayMS   = 200
	pollIntervalMS     = 1000
	errorRetryDelayMS  = 5000
)

type SSEHandler struct {
	l *zap.Logger
}

func NewSSEHandler(l *zap.Logger) *SSEHandler {
	return &SSEHandler{l: l}
}

func (h *SSEHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/log-streams/:logStreamId/logs/sse", h.StreamLogs)
	return nil
}

func (h *SSEHandler) StreamLogs(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	logStreamID := c.Param("logStreamId")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		respondError(c, http.StatusInternalServerError, fmt.Errorf("streaming not supported"))
		return
	}

	ctx := c.Request.Context()
	currentOffset := ""
	isCatchingUp := false
	hasSeenFirstBatch := false

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		logs, err := client.LogStreamReadLogs(ctx, logStreamID, currentOffset)
		if err != nil {
			errData, _ := json.Marshal(map[string]string{"error": "Polling failed"})
			fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", errData)
			flusher.Flush()
			time.Sleep(time.Duration(errorRetryDelayMS) * time.Millisecond)
			continue
		}

		if len(logs) > 0 {
			if !hasSeenFirstBatch {
				isCatchingUp = len(logs) >= streamingThreshold
				hasSeenFirstBatch = true
			}

			if isCatchingUp {
				data, _ := json.Marshal(logs)
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				flusher.Flush()

				// TODO: check for x-nuon-api-next header when SDK supports it
				// For now, stop catching up when we get fewer logs than threshold
				if len(logs) < streamingThreshold {
					isCatchingUp = false
				}
			} else {
				for _, log := range logs {
					data, _ := json.Marshal([]any{log})
					fmt.Fprintf(c.Writer, "data: %s\n\n", data)
					flusher.Flush()
					time.Sleep(time.Duration(streamingDelayMS) * time.Millisecond)
				}
			}
		}

		time.Sleep(time.Duration(pollIntervalMS) * time.Millisecond)
	}
}
