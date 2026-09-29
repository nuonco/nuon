package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	tailMaxWait             = 30 * time.Second
	tailMinBackoff          = 250 * time.Millisecond
	tailMaxBackoff          = 1 * time.Second
	tailMaxConcurrentProbes = 50

	tailPageSize = 100
)

// why: tailProbeSem bounds in-flight CH probes process-wide; sleeping
// long-pollers don't hold a slot, only callers actively querying CH.
var tailProbeSem = make(chan struct{}, tailMaxConcurrentProbes)

const (
	metricTailProbe            = "log_tail.probe"
	metricTailProbeError       = "log_tail.probe_error"
	metricTailEmptyProbeMs     = "log_tail.empty_probe_ms"
	metricTailHotProbeMs       = "log_tail.hot_probe_ms"
	metricTailOutcome          = "log_tail.outcome"
	metricTailSessionMs        = "log_tail.session_ms"
	metricTailProbesPerSession = "log_tail.probes_per_session"
	metricTailRowsPerResponse  = "log_tail.rows_per_response"
)

const (
	tailOutcomeHotHit       = "hot_hit"
	tailOutcomeIdleThenHit  = "idle_then_hit"
	tailOutcomeTimeoutEmpty = "timeout_empty"
	tailOutcomeClientCancel = "client_cancel"
	tailOutcomeError        = "error"
)

// LogStreamTailLogsResponse is the wire shape returned by the tail endpoint.
// `next` is the composite cursor (`<unix_nano>:<id>`) the caller should send
// on its next request — empty means "no rows yet, reuse your previous cursor".
// `has_more` is true when the probe returned a full page; the caller should
// re-request immediately to drain the backlog instead of long-polling.
type LogStreamTailLogsResponse struct {
	Logs    []app.OtelLogRecord `json:"logs"`
	Next    string              `json:"next"`
	HasMore bool                `json:"has_more"`
}

// @ID						LogStreamTailLogs
// @Summary				long-poll tail a log stream
// @Description			Returns rows after the supplied composite cursor, long-polling up to ~30s for new rows on an idle stream.
// @Param					log_stream_id	path	string	true	"log stream ID"
// @Param					since			query	string	false	"composite cursor in the form `<unix_nano>:<id>`; empty starts from the oldest row"
// @Param					wait			query	string	false	"max wait for new rows (Go duration, capped server-side at 30s)"
// @Param					start_time			query	string		false	"only return records with timestamp >= start_time (RFC3339)"
// @Param					end_time			query	string		false	"only return records with timestamp <= end_time (RFC3339)"
// @Param					service_name		query	[]string	false	"filter by service_name (repeatable) collectionFormat(multi)"
// @Param					scope_name			query	[]string	false	"filter by scope_name (repeatable; e.g. oteljob, system) collectionFormat(multi)"
// @Param					scope_version		query	[]string	false	"filter by scope_version (repeatable) collectionFormat(multi)"
// @Param					resource_schema_url	query	[]string	false	"filter by resource_schema_url (repeatable) collectionFormat(multi)"
// @Param					scope_schema_url	query	[]string	false	"filter by scope_schema_url (repeatable) collectionFormat(multi)"
// @Param					severity_text		query	[]string	false	"filter by severity_text (repeatable; INFO/WARN/ERROR/...) collectionFormat(multi)"
// @Param					severity_number_min	query	int			false	"filter by severity_number >= N (OTEL: TRACE=1..FATAL=24)"
// @Param					severity_number_max	query	int			false	"filter by severity_number <= N (OTEL: TRACE=1..FATAL=24)"
// @Param					trace_id			query	string		false	"filter by exact trace_id (dedicated CH column)"
// @Param					span_id				query	string		false	"filter by exact span_id (dedicated CH column)"
// @Param					trace_flags			query	int			false	"filter by exact trace_flags (UInt8)"
// @Param					runner_id			query	string		false	"filter by runner_id"
// @Param					runner_job_id		query	string		false	"filter by runner_job_id (part of CH ORDER BY — efficient)"
// @Param					runner_group_id		query	string		false	"filter by runner_group_id"
// @Param					runner_job_execution_id		query	string	false	"filter by runner_job_execution_id"
// @Param					runner_job_execution_step	query	string	false	"filter by runner_job_execution_step"
// @Param					tool				query	[]string	false	"filter by log_attributes['nuon.tool'] (repeatable; e.g. helm, terraform, kubernetes_manifest, runner) collectionFormat(multi)"
// @Param					helm_release_name	query	string		false	"filter by log_attributes['helm.release_name']"
// @Param					helm_chart_name		query	string		false	"filter by log_attributes['helm.chart_name']"
// @Param					helm_chart_id		query	string		false	"filter by log_attributes['helm.chart_id']"
// @Param					helm_namespace		query	string		false	"filter by log_attributes['helm.namespace']"
// @Param					helm_operation		query	string		false	"filter by log_attributes['helm.operation']"
// @Param					tf_workspace_id		query	string		false	"filter by log_attributes['tf.workspace_id']"
// @Param					tf_operation		query	string		false	"filter by log_attributes['tf.operation']"
// @Param					k8s_kind			query	string		false	"filter by log_attributes['k8s.kind']"
// @Param					k8s_namespace		query	string		false	"filter by log_attributes['k8s.namespace']"
// @Param					k8s_name			query	string		false	"filter by log_attributes['k8s.name']"
// @Param					k8s_operation		query	string		false	"filter by log_attributes['k8s.operation']"
// @Param					attr				query	[]string	false	"generic log_attributes filter as 'key:value' (value must be non-empty; repeatable, max 16 across all attr params) collectionFormat(multi)"
// @Param					resource_attr		query	[]string	false	"generic resource_attributes filter as 'key:value' (value must be non-empty; repeatable, max 16 across all attr params) collectionFormat(multi)"
// @Param					scope_attr			query	[]string	false	"generic scope_attributes filter as 'key:value' (value must be non-empty; repeatable, max 16 across all attr params) collectionFormat(multi)"
// @Param					q					query	string		false	"case-insensitive substring filter on log body"
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
// @Failure				503	{object}	stderr.ErrResponse
// @Success				200	{object}	LogStreamTailLogsResponse
// @Router					/v1/log-streams/{log_stream_id}/logs/tail [GET]
func (s *service) LogStreamTailLogs(ctx *gin.Context) {
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

	cursor, err := parseTailCursor(ctx.Query("since"))
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(errors.Wrap(err, "invalid `since` cursor")))
		return
	}

	filters, err := parseLogFilters(ctx)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	wait := tailMaxWait
	if w := ctx.Query("wait"); w != "" {
		d, err := time.ParseDuration(w)
		if err != nil {
			ctx.Error(stderr.NewInvalidRequest(errors.Wrap(err, "invalid `wait` duration")))
			return
		}
		if d > 0 && d < wait {
			wait = d
		}
	}

	startedAt := time.Now()
	deadline := startedAt.Add(wait)

	backoff := tailMinBackoff
	firstIter := true
	probes := 0
	for {
		probeStart := time.Now()
		logs, next, hasMore, qerr := s.tailProbe(ctx.Request.Context(), orgID, logStreamID, cursor, filters)
		probes++
		s.mw.Count(metricTailProbe, 1, nil)
		if qerr != nil {
			s.mw.Count(metricTailProbeError, 1, nil)
			if ctx.Request.Context().Err() != nil {
				s.emitTailExit(tailOutcomeClientCancel, startedAt, probes)
				return
			}
			if !isTransientTailProbeError(qerr) {
				s.emitTailExit(tailOutcomeError, startedAt, probes)
				ctx.Error(errors.Wrap(qerr, "unable to probe log tail"))
				return
			}

			firstIter = false
			remaining := time.Until(deadline)
			if remaining <= 0 {
				s.emitTailExit(tailOutcomeError, startedAt, probes)
				s.l.Warn("log tail probe retries exhausted", zap.Error(qerr))
				writeTailUnavailable(ctx)
				return
			}

			sleep := backoff + jitter(backoff)
			if sleep > remaining {
				sleep = remaining
			}
			select {
			case <-ctx.Request.Context().Done():
				s.emitTailExit(tailOutcomeClientCancel, startedAt, probes)
				return
			case <-time.After(sleep):
			}
			if time.Until(deadline) <= 0 {
				s.emitTailExit(tailOutcomeError, startedAt, probes)
				s.l.Warn("log tail probe retries exhausted", zap.Error(qerr))
				writeTailUnavailable(ctx)
				return
			}
			backoff *= 2
			if backoff > tailMaxBackoff {
				backoff = tailMaxBackoff
			}
			continue
		}

		if len(logs) > 0 {
			outcome := tailOutcomeIdleThenHit
			if firstIter {
				s.mw.Timing(metricTailHotProbeMs, time.Since(probeStart), nil)
				outcome = tailOutcomeHotHit
			}
			s.emitTailExit(outcome, startedAt, probes)
			s.mw.Distribution(metricTailRowsPerResponse, float64(len(logs)), nil)
			ctx.JSON(http.StatusOK, LogStreamTailLogsResponse{
				Logs:    logs,
				Next:    next,
				HasMore: hasMore,
			})
			return
		}

		s.mw.Timing(metricTailEmptyProbeMs, time.Since(probeStart), nil)
		firstIter = false

		if next != "" {
			if cur, perr := parseTailCursor(next); perr == nil {
				cursor = cur
			}
		}

		select {
		case <-ctx.Request.Context().Done():
			s.emitTailExit(tailOutcomeClientCancel, startedAt, probes)
			return
		default:
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			s.emitTailExit(tailOutcomeTimeoutEmpty, startedAt, probes)
			s.mw.Distribution(metricTailRowsPerResponse, 0, nil)
			ctx.JSON(http.StatusOK, LogStreamTailLogsResponse{
				Logs:    []app.OtelLogRecord{},
				Next:    encodeTailCursor(cursor),
				HasMore: false,
			})
			return
		}

		sleep := backoff + jitter(backoff)
		if sleep > remaining {
			sleep = remaining
		}
		select {
		case <-ctx.Request.Context().Done():
			s.emitTailExit(tailOutcomeClientCancel, startedAt, probes)
			return
		case <-time.After(sleep):
		}

		backoff *= 2
		if backoff > tailMaxBackoff {
			backoff = tailMaxBackoff
		}
	}
}

func (s *service) emitTailExit(result string, startedAt time.Time, probes int) {
	tags := []string{"result:" + result}
	s.mw.Count(metricTailOutcome, 1, tags)
	s.mw.Timing(metricTailSessionMs, time.Since(startedAt), tags)
	s.mw.Distribution(metricTailProbesPerSession, float64(probes), tags)
}

type tailCursor struct {
	tsNano int64
	id     string
}

func parseTailCursor(raw string) (tailCursor, error) {
	if raw == "" {
		return tailCursor{}, nil
	}
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return tailCursor{}, errors.New("expected `<unix_nano>:<id>`")
	}
	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return tailCursor{}, errors.Wrap(err, "unable to parse timestamp")
	}
	return tailCursor{tsNano: ts, id: parts[1]}, nil
}

func encodeTailCursor(c tailCursor) string {
	if c.tsNano == 0 && c.id == "" {
		return ""
	}
	return fmt.Sprintf("%d:%s", c.tsNano, c.id)
}

func (s *service) tailProbe(parent context.Context, orgID, logStreamID string, cursor tailCursor, filters logFilters) ([]app.OtelLogRecord, string, bool, error) {
	select {
	case tailProbeSem <- struct{}{}:
	case <-parent.Done():
		return nil, "", false, parent.Err()
	}
	defer func() { <-tailProbeSem }()

	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	q := s.chDB.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID)

	// why: `timestamp` stays unwrapped on the left so the CH sort key on
	// (org_id, log_stream_id, runner_job_id, timestamp_time,
	// timestamp) can prune granules.
	//
	// When the caller carries an id, use a strictly-greater
	// composite cursor so rows sharing a timestamp paginate without
	// dupes. When the caller hands off from the legacy read
	// endpoint (which only knows the timestamp), id is empty and
	// the safe interpretation is "strictly after this ns" — the
	// legacy paginator already consumed everything at the boundary.
	addCursorPred := func(q *gorm.DB) *gorm.DB {
		if cursor.tsNano <= 0 {
			return q
		}
		if cursor.id != "" {
			return q.Where(
				"(timestamp > fromUnixTimestamp64Nano(?)) OR (timestamp = fromUnixTimestamp64Nano(?) AND id > ?)",
				cursor.tsNano, cursor.tsNano, cursor.id,
			)
		}
		return q.Where("timestamp > fromUnixTimestamp64Nano(?)", cursor.tsNano)
	}

	q = applyLogFilters(addCursorPred(q), filters)

	hwm := tailCursor{}
	if hasAnyFilter(filters) {
		var err error
		hwm, err = s.tailHighWaterMark(ctx, orgID, logStreamID, cursor)
		if err != nil {
			return nil, "", false, errors.Wrap(tailProbeQueryError(ctx, err), "unable to query log tail high-water mark")
		}
		if hwm.tsNano == 0 {
			return nil, "", false, nil
		}
		q = q.Where(
			"(timestamp < fromUnixTimestamp64Nano(?)) OR (timestamp = fromUnixTimestamp64Nano(?) AND id <= ?)",
			hwm.tsNano, hwm.tsNano, hwm.id,
		)
	}

	var rows []app.OtelLogRecord
	res := q.Order("timestamp ASC, id ASC").
		Limit(tailPageSize + 1).
		Find(&rows)
	if res.Error != nil {
		return nil, "", false, errors.Wrap(tailProbeQueryError(ctx, res.Error), "unable to query log tail")
	}

	hasMore := len(rows) > tailPageSize
	if hasMore {
		rows = rows[:tailPageSize]
	}

	if len(rows) == 0 {
		if hwm.tsNano > 0 {
			return rows, encodeTailCursor(hwm), false, nil
		}
		return rows, "", false, nil
	}
	last := rows[len(rows)-1]
	next := encodeTailCursor(tailCursor{tsNano: last.Timestamp.UnixNano(), id: last.ID})
	return rows, next, hasMore, nil
}

func (s *service) tailHighWaterMark(ctx context.Context, orgID, logStreamID string, cursor tailCursor) (tailCursor, error) {
	q := "SELECT timestamp AS ts, id FROM otel_log_records WHERE org_id = ? AND log_stream_id = ?"
	args := []interface{}{orgID, logStreamID}

	if cursor.tsNano > 0 {
		if cursor.id != "" {
			q += " AND ((timestamp > fromUnixTimestamp64Nano(?)) OR (timestamp = fromUnixTimestamp64Nano(?) AND id > ?))"
			args = append(args, cursor.tsNano, cursor.tsNano, cursor.id)
		} else {
			q += " AND timestamp > fromUnixTimestamp64Nano(?)"
			args = append(args, cursor.tsNano)
		}
	}
	q += " ORDER BY timestamp DESC, id DESC LIMIT 1"

	var row struct {
		Ts time.Time `gorm:"column:ts"`
		ID string    `gorm:"column:id"`
	}
	if err := s.chDB.WithContext(ctx).Raw(q, args...).Scan(&row).Error; err != nil {
		return tailCursor{}, err
	}
	if row.ID == "" || row.Ts.IsZero() || row.Ts.UnixNano() <= 0 {
		return tailCursor{}, nil
	}
	return tailCursor{tsNano: row.Ts.UnixNano(), id: row.ID}, nil
}

func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	half := int64(d) / 2
	return time.Duration(rand.Int63n(half) - half/2)
}
