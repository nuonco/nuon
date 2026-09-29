package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// LogStreamSpan is the wire-format span returned by LogStreamReadSpans.
//
// One row per span. The frontend assembles the call tree from
// parent_span_id (empty string for the root). Ordered by start_timestamp
// ASC so a stable parent-before-child traversal usually works as a fast
// path; the frontend should still tolerate any order.
type LogStreamSpan struct {
	SpanID         string    `json:"span_id"`
	ParentSpanID   string    `json:"parent_span_id"`
	TraceID        string    `json:"trace_id"`
	SpanName       string    `json:"span_name"`
	SpanKind       string    `json:"span_kind"`
	ServiceName    string    `json:"service_name"`
	ScopeName      string    `json:"scope_name"`
	StatusCode     string    `json:"status_code"`
	StatusMessage  string    `json:"status_message"`
	StartTimestamp time.Time `json:"start_timestamp"`
	EndTimestamp   time.Time `json:"end_timestamp"`
	// DurationNs is the wall-clock span duration in nanoseconds, copied
	// straight from the otel_traces.duration column.
	DurationNs int64 `json:"duration_ns"`
	// Per-execution identifiers extracted server-side at ingest time.
	// Empty if the span was emitted outside a job context.
	RunnerJobID            string            `json:"runner_job_id"`
	RunnerJobExecutionID   string            `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string            `json:"runner_job_execution_step"`
	Attributes             map[string]string `json:"attributes"`
}

// @ID						LogStreamReadSpans
// @Summary				read a log stream's trace spans
// @Description.markdown	log_stream_read_spans.md
// @Param					log_stream_id	path	string	true	"log stream ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]LogStreamSpan
// @Router					/v1/log-streams/{log_stream_id}/spans [GET]
func (s *service) LogStreamReadSpans(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read org id from context"))
		return
	}

	if _, err := s.getOrgLogStream(ctx, logStreamID, orgID); err != nil {
		ctx.Error(errors.Wrap(err, "unable to get log stream"))
		return
	}

	spans, err := s.getLogStreamSpans(ctx, logStreamID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read log stream spans"))
		return
	}

	ctx.JSON(http.StatusOK, spans)
}

func (s *service) getLogStreamSpans(ctx context.Context, logStreamID string) ([]LogStreamSpan, error) {
	ctx, cancelFn := context.WithTimeout(ctx, time.Second*5)
	defer cancelFn()

	var jobIDs []string
	res := s.db.WithContext(ctx).
		Model(&app.RunnerJob{}).
		Where("log_stream_id = ?", logStreamID).
		Pluck("id", &jobIDs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to resolve log stream to runner jobs")
	}
	if len(jobIDs) == 0 {
		return []LogStreamSpan{}, nil
	}

	var rows []app.OtelTraceIngestion
	res = s.chDB.WithContext(ctx).
		Where("runner_job_id IN ?", jobIDs).
		Order("timestamp ASC").
		Find(&rows)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to read trace spans")
	}

	spans := make([]LogStreamSpan, 0, len(rows))
	for _, r := range rows {
		spans = append(spans, LogStreamSpan{
			SpanID:                 r.SpanID,
			ParentSpanID:           r.ParentSpanID,
			TraceID:                r.TraceID,
			SpanName:               r.SpanName,
			SpanKind:               r.SpanKind,
			ServiceName:            r.ServiceName,
			ScopeName:              r.ScopeName,
			StatusCode:             r.StatusCode,
			StatusMessage:          r.StatusMessage,
			StartTimestamp:         r.Timestamp,
			EndTimestamp:           r.Timestamp.Add(time.Duration(r.Duration)),
			DurationNs:             r.Duration,
			RunnerJobID:            r.RunnerJobID,
			RunnerJobExecutionID:   r.RunnerJobExecutionID,
			RunnerJobExecutionStep: r.RunnerJobExecutionStep,
			Attributes:             r.SpanAttributes,
		})
	}

	return spans, nil
}
