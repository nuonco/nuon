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

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// Long-poll tail tuning. These are intentionally not user-configurable.
const (
	// Maximum time the handler will hold the request open before returning
	// an empty response. The dashboard BFF re-issues the request immediately
	// after receiving an empty payload, so this is the worst-case dead time
	// between a runner emitting a log and the user seeing it during idle.
	tailMaxWait = 30 * time.Second
	// Floor / ceiling for the per-handler backoff between empty probes.
	// 250ms is the user-visible first-line latency on the happy path
	// (rows already buffered by ClickHouse async insert); 1s is the
	// idle-stream rate we settle into.
	tailMinBackoff = 250 * time.Millisecond
	tailMaxBackoff = 1 * time.Second
	// Per-pod cap on concurrent ClickHouse probes. Sleeping long-poll
	// handlers don't count against this — only the moment they're
	// actively querying CH. 50 is comfortably under the default CH
	// query concurrency and the eventual SetMaxOpenConns(50) pool size.
	tailMaxConcurrentProbes = 50

	tailPageSize = 100
)

// tailProbeSem bounds in-flight CH probes process-wide; sleeping
// long-pollers don't hold a slot, only callers actively querying CH.
var tailProbeSem = make(chan struct{}, tailMaxConcurrentProbes)

// Metric names. All three are tagged with `org_id` so per-org rollout
// (per the feature flag) shows up clearly on dashboards.
const (
	// metricTailProbe — how many CH probes is this feature generating
	// per org? Cost / load guardrail. The 250ms→1s expo backoff puts
	// the steady-state per-idle-subscriber rate at ~1 qps; if real
	// numbers run materially hotter (e.g. rows trickle in just often
	// enough to keep resetting the backoff) this is where it shows
	// before CH does.
	metricTailProbe = "log_tail.probe"
	// metricTailEmptyProbeMs — when CH had nothing to return, how
	// long did the round-trip cost? Health of the idle path. Should
	// sit in single-digit ms; a creeping p95 means CH is doing real
	// work on empty probes (sort-key pruning regressed, async_insert
	// backed up). Early-warning metric.
	metricTailEmptyProbeMs = "log_tail.empty_probe_ms"
	// metricTailFirstByteMs — the actual UX metric. From the moment
	// the BFF issued the request, how long until the user saw a line
	// of output? On a live job this should sit near tailMinBackoff
	// (250ms) + the CH read. If p50 isn't materially better than the
	// legacy 1s-poll path, the feature isn't doing its job and we
	// should reconsider before broader rollout.
	metricTailFirstByteMs = "log_tail.first_byte_ms"
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
// @Description			Returns rows after the supplied composite cursor, long-polling up to ~30s for new rows on an idle stream. Behind the `log-tail-long-poll` org feature flag.
// @Param					log_stream_id	path	string	true	"log stream ID"
// @Param					since			query	string	false	"composite cursor in the form `<unix_nano>:<id>`; empty starts from the oldest row"
// @Param					wait			query	string	false	"max wait for new rows (Go duration, capped server-side at 30s)"
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
// @Success				200	{object}	LogStreamTailLogsResponse
// @Router					/v1/log-streams/{log_stream_id}/logs/tail [GET]
func (s *service) LogStreamTailLogs(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read org id from context"))
		return
	}

	// Feature gate — when off, return 404 so callers (BFF / SDK) can
	// treat the tail endpoint as if it doesn't exist for this org and
	// fall back to the legacy polling read path. 403/501 would force
	// callers to surface this as a real error.
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureLogTailLongPoll)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to check log-tail-long-poll feature"))
		return
	}
	if !enabled {
		ctx.AbortWithStatus(http.StatusNotFound)
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

	tags := metrics.ToTags(map[string]string{"org_id": orgID})

	startedAt := time.Now()
	deadline := startedAt.Add(wait)

	backoff := tailMinBackoff
	for {
		probeStart := time.Now()
		logs, next, hasMore, qerr := s.tailProbe(ctx.Request.Context(), orgID, logStreamID, cursor)
		s.mw.Count(metricTailProbe, 1, tags)
		if qerr != nil {
			ctx.Error(errors.Wrap(qerr, "unable to probe log tail"))
			return
		}

		if len(logs) > 0 {
			s.mw.Timing(metricTailFirstByteMs, time.Since(startedAt), tags)
			ctx.JSON(http.StatusOK, LogStreamTailLogsResponse{
				Logs:    logs,
				Next:    next,
				HasMore: hasMore,
			})
			return
		}

		s.mw.Timing(metricTailEmptyProbeMs, time.Since(probeStart), tags)

		// Drop out at the wait deadline (or when the client closes the
		// connection). Returning an empty payload is the long-poll
		// equivalent of "still nothing, ask again".
		select {
		case <-ctx.Request.Context().Done():
			return
		default:
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
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
			return
		case <-time.After(sleep):
		}

		// Expo backoff capped at tailMaxBackoff so the steady-state
		// idle stream costs ~1 CH probe / sec / subscriber.
		backoff *= 2
		if backoff > tailMaxBackoff {
			backoff = tailMaxBackoff
		}
	}
}

// tailCursor encodes the (timestamp, id) tiebreak used to paginate
// monotonically across rows that share a timestamp. ClickHouse stores
// the timestamp at nanosecond precision but the runner can emit multiple
// rows in the same nanosecond, so an id tiebreaker is required for a
// correct strictly-greater cursor.
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

// tailProbe performs a single bounded ClickHouse query for rows after the
// supplied cursor. It blocks on the per-pod semaphore so a burst of tail
// requests doesn't translate to an unbounded burst of CH connections.
func (s *service) tailProbe(parent context.Context, orgID, logStreamID string, cursor tailCursor) ([]app.OtelLogRecord, string, bool, error) {
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

	if cursor.tsNano > 0 {
		// `timestamp` stays unwrapped on the left so the CH sort key on
		// (org_id, log_stream_id, runner_job_id, timestamp_time,
		// timestamp) can prune granules.
		//
		// When the caller carries an id, use a strictly-greater
		// composite cursor so rows sharing a timestamp paginate without
		// dupes. When the caller hands off from the legacy read
		// endpoint (which only knows the timestamp), id is empty and
		// the safe interpretation is "strictly after this ns" — the
		// legacy paginator already consumed everything at the boundary.
		if cursor.id != "" {
			q = q.Where(
				"(timestamp > fromUnixTimestamp64Nano(?)) OR (timestamp = fromUnixTimestamp64Nano(?) AND id > ?)",
				cursor.tsNano, cursor.tsNano, cursor.id,
			)
		} else {
			q = q.Where("timestamp > fromUnixTimestamp64Nano(?)", cursor.tsNano)
		}
	}

	// LIMIT pageSize+1 so we can report `has_more` without a separate
	// COUNT(*). The legacy read endpoint does a COUNT(*) per request,
	// which is what we're explicitly avoiding here.
	var rows []app.OtelLogRecord
	res := q.Order("timestamp ASC, id ASC").
		Limit(tailPageSize + 1).
		Find(&rows)
	if res.Error != nil {
		return nil, "", false, errors.Wrap(res.Error, "unable to query log tail")
	}

	hasMore := len(rows) > tailPageSize
	if hasMore {
		rows = rows[:tailPageSize]
	}

	if len(rows) == 0 {
		return rows, "", false, nil
	}
	last := rows[len(rows)-1]
	next := encodeTailCursor(tailCursor{tsNano: last.Timestamp.UnixNano(), id: last.ID})
	return rows, next, hasMore, nil
}

// jitter returns a uniformly distributed value in [-d/4, +d/4) so a
// thundering herd of long-pollers don't synchronize their probes.
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	half := int64(d) / 2
	return time.Duration(rand.Int63n(half) - half/2)
}
