package handlers

import (
	"context"
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
	// pollInterval is the legacy 1s-poll cadence, used for DESC sessions.
	pollInterval    = 1 * time.Second
	errorRetryDelay = 5 * time.Second
	// streamStatusCheck bounds how long we'll sit on an open long-poll
	// before re-asking the server whether the stream has closed.
	streamStatusCheck = 10 * time.Second

	// tailInitialWait is the first long-poll probe's wait override. The
	// server default is ~30s; we want the very first probe to return
	// quickly so a completed-and-empty stream emits "complete" before
	// the user perceives a stall on page load.
	tailInitialWait = "1s"
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

type streamSession struct {
	client      nuon.Client
	l           *zap.Logger
	logStreamID string
	runnerJobID string
	order       string
	isOpen      bool

	hasSeenFirstBatch bool
	isCatchingUp      bool
	lastStatusCheck   time.Time

	sendEvent  func(logs []*models.AppOtelLogRecord)
	sendStatus func(status string)
	sendError  func(msg string)
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

	ctx := c.Request.Context()

	order := c.Query("order")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	logStream, err := client.GetLogStream(ctx, logStreamID)
	if err != nil {
		h.l.Error("failed to get log stream", zap.Error(err), zap.String("logStreamID", logStreamID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get log stream"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	sess := &streamSession{
		client:          client,
		l:               h.l,
		logStreamID:     logStreamID,
		runnerJobID:     c.Query("runner_job_id"),
		order:           order,
		isOpen:          logStream.Open,
		lastStatusCheck: time.Now(),
		sendEvent: func(logs []*models.AppOtelLogRecord) {
			b, _ := json.Marshal(logs)
			fmt.Fprintf(c.Writer, "data: %s\n\n", b)
			c.Writer.Flush()
		},
		sendStatus: func(status string) {
			fmt.Fprintf(c.Writer, "event: status\ndata: %s\n\n", status)
			c.Writer.Flush()
		},
		sendError: func(msg string) {
			b, _ := json.Marshal(map[string]string{"error": msg})
			fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", b)
			c.Writer.Flush()
		},
	}

	// Tail endpoint only supports ASC; DESC stays on the legacy path.
	if order == "asc" {
		sess.runTail(ctx)
		return
	}
	sess.runLegacy(ctx, "")
}

// runTail drives the long-poll tail endpoint. Transient errors retry on
// the tail path rather than swapping to the legacy poller mid-session —
// the org opted into the new path, keep them on it.
func (s *streamSession) runTail(ctx context.Context) {
	since := "" // empty cursor: server starts at the oldest row and drains via has_more.
	wait := tailInitialWait

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := s.client.LogStreamTailLogs(ctx, s.logStreamID, since, wait)
		if err != nil {
			s.l.Warn("log tail poll failed",
				zap.String("logStreamID", s.logStreamID),
				zap.Error(err))
			s.sendError("Polling failed")
			select {
			case <-ctx.Done():
				return
			case <-time.After(errorRetryDelay):
			}
			continue
		}

		// After the first probe, drop the short initial-wait override so
		// idle steady-state is bounded by the server's 30s wait cap.
		wait = ""

		logs := filterLogsByRunnerJobID(resp.Logs, s.runnerJobID)
		if len(logs) > 0 {
			if !s.hasSeenFirstBatch {
				s.isCatchingUp = len(logs) >= streamingThreshold || resp.HasMore
				s.hasSeenFirstBatch = true
				if s.isCatchingUp {
					s.sendStatus("catching-up")
				}
			}

			if s.isCatchingUp {
				s.sendEvent(logs)
			} else if !s.isOpen {
				// Closed stream: no typewriter pacing, just dump.
				s.sendEvent(logs)
			} else {
				// Live stream: pace one log at a time so output streams
				// in rather than landing in 100-line jumps.
				for _, log := range logs {
					select {
					case <-ctx.Done():
						return
					case <-time.After(streamingDelay):
					}
					s.sendEvent([]*models.AppOtelLogRecord{log})
				}
			}
		}
		if s.isCatchingUp && !resp.HasMore {
			s.isCatchingUp = false
			s.sendStatus("live")
		}
		if resp.Next != "" {
			since = resp.Next
		}

		if !s.isOpen && len(resp.Logs) == 0 {
			s.sendStatus("complete")
			<-ctx.Done()
			return
		}

		// Re-check the stream's open state — without this we'd sit on
		// the long-poll forever after a job finishes and stops emitting.
		// When we discover the stream just closed, drop the next probe's
		// wait window so we drain and emit "complete" promptly rather
		// than blocking another full 30s on the server's default wait.
		if s.isOpen && time.Since(s.lastStatusCheck) >= streamStatusCheck {
			s.lastStatusCheck = time.Now()
			ls, err := s.client.GetLogStream(ctx, s.logStreamID)
			if err == nil && !ls.Open {
				s.isOpen = false
				wait = tailInitialWait
			}
		}
	}
}

func (s *streamSession) runLegacy(ctx context.Context, currentOffset string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		logs, nextOffset, err := s.client.LogStreamReadLogsWithNextOffset(ctx, s.logStreamID, currentOffset, s.order)
		if err != nil {
			s.sendError("Polling failed")
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

		logs = filterLogsByRunnerJobID(logs, s.runnerJobID)
		paginationComplete := nextOffset == ""
		if len(logs) > 0 {
			if !s.hasSeenFirstBatch {
				s.isCatchingUp = len(logs) >= streamingThreshold
				s.hasSeenFirstBatch = true
				if s.isCatchingUp {
					s.sendStatus("catching-up")
				}
			}

			if s.isCatchingUp {
				s.sendEvent(logs)
				if !paginationComplete {
					continue
				}
			} else if !s.isOpen {
				s.sendEvent(logs)
			} else {
				for _, log := range logs {
					select {
					case <-ctx.Done():
						return
					case <-time.After(streamingDelay):
					}
					s.sendEvent([]*models.AppOtelLogRecord{log})
				}
			}
		}
		if s.isCatchingUp && paginationComplete {
			s.isCatchingUp = false
			s.sendStatus("live")
		}

		if !s.isOpen && nextOffset == "" {
			s.sendStatus("complete")
			<-ctx.Done()
			return
		}

		if s.isOpen && time.Since(s.lastStatusCheck) >= streamStatusCheck {
			s.lastStatusCheck = time.Now()
			ls, err := s.client.GetLogStream(ctx, s.logStreamID)
			if err == nil && !ls.Open {
				s.isOpen = false
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

	order := c.Query("order")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="logs-%s.txt"`, logStreamID))
	c.Writer.WriteHeader(http.StatusOK)

	jobOutputOnly := c.Query("job_output") == "true"
	runnerJobID := c.Query("runner_job_id")

	for {
		logs, nextOffset, err := client.LogStreamReadLogsWithNextOffset(ctx, logStreamID, offset, order)
		if err != nil {
			h.l.Error("failed to read log stream logs", zap.Error(err), zap.String("logStreamID", logStreamID))
			break
		}

		batchSize := len(logs)
		logs = filterLogsByRunnerJobID(logs, runnerJobID)
		for _, log := range logs {
			if jobOutputOnly && log.ScopeName != "oteljob" {
				continue
			}
			fmt.Fprintf(c.Writer, "[%s] [%s] [%s] %s\n", log.Timestamp, log.SeverityText, log.ServiceName, log.Body)
		}
		c.Writer.Flush()

		totalLogs += batchSize
		if nextOffset == "" || totalLogs >= maxDownloadLogs {
			break
		}
		offset = nextOffset
	}
}

func filterLogsByRunnerJobID(logs []*models.AppOtelLogRecord, runnerJobID string) []*models.AppOtelLogRecord {
	if runnerJobID == "" {
		return logs
	}

	filtered := make([]*models.AppOtelLogRecord, 0, len(logs))
	for _, log := range logs {
		if log.RunnerJobID == runnerJobID {
			filtered = append(filtered, log)
		}
	}
	return filtered
}
