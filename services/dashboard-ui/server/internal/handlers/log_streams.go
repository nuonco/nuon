package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
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

// parseLogFiltersFromQuery maps the browser's filter query params onto the
// SDK filter struct sent to ctl-api. It's an allowlist — `since`, `wait` and
// `order` are parsed separately and never flow through here. Malformed
// numeric params are skipped; ctl-api re-validates everything.
func parseLogFiltersFromQuery(c *gin.Context) *nuon.LogStreamLogFilters {
	q := c.Request.URL.Query()
	if len(q) == 0 {
		return nil
	}

	f := &nuon.LogStreamLogFilters{
		StartTime:              q.Get("start_time"),
		EndTime:                q.Get("end_time"),
		ServiceNames:           q["service_name"],
		ScopeNames:             q["scope_name"],
		ScopeVersions:          q["scope_version"],
		SeverityTexts:          q["severity_text"],
		ResourceSchemaURLs:     q["resource_schema_url"],
		ScopeSchemaURLs:        q["scope_schema_url"],
		TraceID:                q.Get("trace_id"),
		SpanID:                 q.Get("span_id"),
		RunnerID:               q.Get("runner_id"),
		RunnerJobID:            q.Get("runner_job_id"),
		RunnerGroupID:          q.Get("runner_group_id"),
		RunnerJobExecutionID:   q.Get("runner_job_execution_id"),
		RunnerJobExecutionStep: q.Get("runner_job_execution_step"),
		Tools:                  q["tool"],
		HelmReleaseName:        q.Get("helm_release_name"),
		HelmChartName:          q.Get("helm_chart_name"),
		HelmChartID:            q.Get("helm_chart_id"),
		HelmNamespace:          q.Get("helm_namespace"),
		HelmOperation:          q.Get("helm_operation"),
		TfWorkspaceID:          q.Get("tf_workspace_id"),
		TfOperation:            q.Get("tf_operation"),
		K8sKind:                q.Get("k8s_kind"),
		K8sNamespace:           q.Get("k8s_namespace"),
		K8sName:                q.Get("k8s_name"),
		K8sOperation:           q.Get("k8s_operation"),
		Attrs:                  q["attr"],
		ResourceAttrs:          q["resource_attr"],
		ScopeAttrs:             q["scope_attr"],
		BodyContains:           q.Get("q"),
	}
	if v, err := strconv.ParseInt(q.Get("severity_number_min"), 10, 64); err == nil {
		f.SeverityNumberMin = v
	}
	if v, err := strconv.ParseInt(q.Get("severity_number_max"), 10, 64); err == nil {
		f.SeverityNumberMax = v
	}
	if v, err := strconv.ParseInt(q.Get("trace_flags"), 10, 64); err == nil {
		f.TraceFlags = v
	}
	if logFiltersEmpty(f) {
		return nil
	}
	return f
}

// ctl-api rejects some filter inputs deterministically (400/401/403/404);
// retrying those with identical params can never succeed, so the session
// ends instead of looping on errorRetryDelay.
type statusCoder interface{ Code() int }

func isTerminalAPIError(err error) bool {
	var sc statusCoder
	if !errors.As(err, &sc) {
		return false
	}
	code := sc.Code()
	return code >= 400 && code < 500
}

func apiErrorMessage(err error) string {
	var br interface {
		GetPayload() *models.StderrErrResponse
	}
	if errors.As(err, &br) {
		if p := br.GetPayload(); p != nil && p.Description != "" {
			return p.Description
		}
	}
	return "Polling failed"
}

func logFiltersEmpty(f *nuon.LogStreamLogFilters) bool {
	rv := reflect.ValueOf(*f)
	for i := 0; i < rv.NumField(); i++ {
		switch fv := rv.Field(i); fv.Kind() {
		case reflect.String:
			if fv.String() != "" {
				return false
			}
		case reflect.Int64:
			if fv.Int() != 0 {
				return false
			}
		case reflect.Slice:
			if fv.Len() > 0 {
				return false
			}
		}
	}
	return true
}

type streamSession struct {
	client      nuon.Client
	l           *zap.Logger
	logStreamID string
	filters     *nuon.LogStreamLogFilters
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
		filters:         parseLogFiltersFromQuery(c),
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

		resp, err := s.client.LogStreamTailLogs(ctx, s.logStreamID, since, wait, s.filters)
		if err != nil {
			s.l.Warn("log tail poll failed",
				zap.String("logStreamID", s.logStreamID),
				zap.Error(err))
			if isTerminalAPIError(err) {
				s.sendError(apiErrorMessage(err))
				return
			}
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

		logs := resp.Logs
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

		logs, nextOffset, err := s.client.LogStreamReadLogsWithNextOffset(ctx, s.logStreamID, currentOffset, s.order, s.filters)
		if err != nil {
			if isTerminalAPIError(err) {
				s.l.Warn("log poll failed terminally",
					zap.String("logStreamID", s.logStreamID),
					zap.Error(err))
				s.sendError(apiErrorMessage(err))
				return
			}
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
	filters := parseLogFiltersFromQuery(c)
	if jobOutputOnly {
		// job_output means user-visible job output only, so it overrides any
		// caller-supplied scope rather than adding to it.
		if filters == nil {
			filters = &nuon.LogStreamLogFilters{}
		}
		filters.ScopeNames = []string{"oteljob"}
	}

	for {
		logs, nextOffset, err := client.LogStreamReadLogsWithNextOffset(ctx, logStreamID, offset, order, filters)
		if err != nil {
			h.l.Error("failed to read log stream logs", zap.Error(err), zap.String("logStreamID", logStreamID))
			break
		}

		for _, log := range logs {
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
