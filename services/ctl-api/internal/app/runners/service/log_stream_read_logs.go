package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	PageSize             int    = 100
	nestedAttributeRegex string = `^(?:[a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)?)$` // https://regex101.com/r/179bxx/1
)

// metricReadFreshnessLagMs — temporary rollout metric to compare the
// flag-off legacy poll path against `log_tail.hot_probe_ms` on the
// flag-on path. For every response with rows, we emit `time.Since(newest
// row timestamp)` — i.e. how stale the freshest row we just returned to
// the user is. On the `mode:poll` slice (BFF 1s-tick live tail) p50
// can't drop below ~500ms by construction because of the polling floor;
// `log_tail.hot_probe_ms` p50 sits in the 300–500ms range with no such
// floor. That gap is the rollout argument expressed in milliseconds.
// Delete once log-tail-long-poll has rolled out broadly enough that we
// no longer need the comparison.
const metricReadFreshnessLagMs = "log_read.freshness_lag_ms"

// readMode classifies a legacy read so the comparison against the tail
// endpoint isn't muddied by search/download/backfill traffic. Only
// `mode:poll` is apples-to-apples with the tail endpoint's role.
const (
	readModePoll     = "poll"     // empty cursor, no filters, ASC — the BFF live-tail loop
	readModeBackfill = "backfill" // non-empty cursor, no filters — paging through history
	readModeFiltered = "filtered" // any filter set, or order=desc — search / ad-hoc
)

func classifyReadMode(cursor int64, order string, f logFilters) string {
	if order != "asc" || hasAnyFilter(f) {
		return readModeFiltered
	}
	if cursor > 0 {
		return readModeBackfill
	}
	return readModePoll
}

func hasAnyFilter(f logFilters) bool {
	return len(f.serviceNames) > 0 ||
		len(f.scopeNames) > 0 ||
		len(f.severityTexts) > 0 ||
		f.tool != "" ||
		f.helmReleaseName != "" ||
		f.helmOperation != "" ||
		f.tfWorkspaceID != "" ||
		f.tfOperation != "" ||
		f.k8sKind != "" ||
		f.k8sNamespace != "" ||
		f.k8sName != "" ||
		f.traceID != "" ||
		f.spanID != "" ||
		f.bodyContains != ""
}

// logFilters holds optional filter values parsed from query parameters.
//
// Top-level columns are filtered with normal SQL clauses. Map columns
// (log_attributes) use ClickHouse's mapContains() in addition to the value
// equality so the bloom_filter skip indexes added in migration 06 can be
// leveraged.
type logFilters struct {
	serviceNames  []string
	scopeNames    []string
	severityTexts []string

	tool            string
	helmReleaseName string
	helmOperation   string
	tfWorkspaceID   string
	tfOperation     string
	k8sKind         string
	k8sNamespace    string
	k8sName         string

	// traceID / spanID are exact-match filters on the dedicated CH columns
	// of the same name. Used by the dashboard span-tree UI to scope logs to
	// a single span node (or the entire trace).
	traceID string
	spanID  string

	bodyContains string
}

func parseLogFilters(ctx *gin.Context) logFilters {
	return logFilters{
		serviceNames:  ctx.QueryArray("service_name"),
		scopeNames:    ctx.QueryArray("scope_name"),
		severityTexts: ctx.QueryArray("severity_text"),

		tool:            ctx.Query("tool"),
		helmReleaseName: ctx.Query("helm_release_name"),
		helmOperation:   ctx.Query("helm_operation"),
		tfWorkspaceID:   ctx.Query("tf_workspace_id"),
		tfOperation:     ctx.Query("tf_operation"),
		k8sKind:         ctx.Query("k8s_kind"),
		k8sNamespace:    ctx.Query("k8s_namespace"),
		k8sName:         ctx.Query("k8s_name"),

		traceID: ctx.Query("trace_id"),
		spanID:  ctx.Query("span_id"),

		bodyContains: firstNonEmpty(ctx.Query("q"), ctx.Query("body_contains")),
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func applyLogFilters(db *gorm.DB, f logFilters) *gorm.DB {
	if len(f.serviceNames) > 0 {
		db = db.Where("service_name IN ?", f.serviceNames)
	}
	if len(f.scopeNames) > 0 {
		db = db.Where("scope_name IN ?", f.scopeNames)
	}
	if len(f.severityTexts) > 0 {
		db = db.Where("severity_text IN ?", f.severityTexts)
	}

	addAttrEq := func(d *gorm.DB, key, val string) *gorm.DB {
		if val == "" {
			return d
		}
		// mapContains() uses the key bloom_filter; equality on the bracket
		// access uses the value bloom_filter. AND'ing both lets ClickHouse
		// prune granules using either index.
		return d.Where("mapContains(log_attributes, ?) AND log_attributes[?] = ?", key, key, val)
	}

	db = addAttrEq(db, "nuon.tool", f.tool)
	db = addAttrEq(db, "helm.release_name", f.helmReleaseName)
	db = addAttrEq(db, "helm.operation", f.helmOperation)
	db = addAttrEq(db, "tf.workspace_id", f.tfWorkspaceID)
	db = addAttrEq(db, "tf.operation", f.tfOperation)
	db = addAttrEq(db, "k8s.kind", f.k8sKind)
	db = addAttrEq(db, "k8s.namespace", f.k8sNamespace)
	db = addAttrEq(db, "k8s.name", f.k8sName)

	// trace_id / span_id are dedicated CH columns populated by the otelzap
	// bridge from the runner's per-op span context (see bins/runner/internal/pkg/op).
	// trace_id has a bloom_filter skip index (see otel_log_record.go); span_id
	// does not — query latency is acceptable today, revisit if it isn't.
	if f.traceID != "" {
		db = db.Where("trace_id = ?", f.traceID)
	}
	if f.spanID != "" {
		db = db.Where("span_id = ?", f.spanID)
	}

	if f.bodyContains != "" {
		db = db.Where("body ILIKE ?", "%"+f.bodyContains+"%")
	}
	return db
}

// @ID						LogStreamReadLogs
// @Summary				read a log stream's logs
// @Description.markdown	log_stream_read_logs.md
// @Param					log_stream_id		path	string	true	"log stream ID"
// @Param					X-Nuon-API-Offset	header	string	false	"log stream offset"
// @Param					order				query	string	false	"sort direction"	default(asc)
// @Param					service_name		query	[]string	false	"filter by service_name (repeatable)"
// @Param					scope_name			query	[]string	false	"filter by scope_name (repeatable)"
// @Param					severity_text		query	[]string	false	"filter by severity_text (repeatable)"
// @Param					tool				query	string		false	"filter by log_attributes['nuon.tool']"
// @Param					helm_release_name	query	string		false	"filter by log_attributes['helm.release_name']"
// @Param					helm_operation		query	string		false	"filter by log_attributes['helm.operation']"
// @Param					tf_workspace_id		query	string		false	"filter by log_attributes['tf.workspace_id']"
// @Param					tf_operation		query	string		false	"filter by log_attributes['tf.operation']"
// @Param					k8s_kind			query	string		false	"filter by log_attributes['k8s.kind']"
// @Param					k8s_namespace		query	string		false	"filter by log_attributes['k8s.namespace']"
// @Param					k8s_name			query	string		false	"filter by log_attributes['k8s.name']"
// @Param					trace_id			query	string		false	"filter by exact trace_id (dedicated CH column)"
// @Param					span_id				query	string		false	"filter by exact span_id (dedicated CH column)"
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
// @Success				200	{object}	[]app.OtelLogRecord
// @Router					/v1/log-streams/{log_stream_id}/logs [GET]
func (s *service) LogStreamReadLogs(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	// Read logs from chDB
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to read org id from context"))
		return
	}

	_, err = s.getOrgLogStream(ctx, logStreamID, orgID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get log stream"))
		return
	}

	// Parse order parameter
	order := ctx.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid order query parameter, must be 'asc' or 'desc'")))
		return
	}

	// Parse cursor
	var cursor int64
	cursorStr := ctx.GetHeader("X-Nuon-API-Offset")
	if cursorStr != "" {
		cursorVal, err := strconv.ParseInt(cursorStr, 10, 64)
		if err != nil {
			ctx.Error(errors.Wrap(err, "unable to parse pagination cursor"))
			return
		}
		cursor = cursorVal
	}

	filters := parseLogFilters(ctx)

	logs, headers, readErr := s.getLogStreamLogs(ctx, logStreamID, orgID, cursor, order, filters)
	if readErr != nil {
		ctx.Error(errors.Wrap(readErr, "unable to read runner logs"))
		return
	}

	// Emit freshness lag of the newest returned row. ASC orders newest
	// last, DESC orders newest first; either way it's the timestamp the
	// user perceives as "most recent visible log" right now.
	if len(logs) > 0 {
		newest := logs[len(logs)-1]
		if order == "desc" {
			newest = logs[0]
		}
		s.mw.Timing(metricReadFreshnessLagMs, time.Since(newest.Timestamp),
			[]string{"mode:" + classifyReadMode(cursor, order, filters)})
	}

	// Set headers
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getLogStreamLogs(ctx context.Context, logStreamID string, orgID string, cursor int64, order string, filters logFilters) ([]app.OtelLogRecord, map[string]string, error) {
	ctx, cancelFn := context.WithTimeout(ctx, time.Second*5)
	defer cancelFn()

	headers := map[string]string{"Range-Units": "items"}

	// Get total count first (filters applied so the count reflects what the
	// caller will actually see for the same query string).
	var totalCount int64
	countQ := s.chDB.WithContext(ctx).
		Model(&app.OtelLogRecord{}).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID)
	countQ = applyLogFilters(countQ, filters)
	if res := countQ.Count(&totalCount); res.Error != nil {
		return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs count")
	}
	headers["count"] = strconv.FormatInt(totalCount, 10)

	// Handle empty results
	if totalCount == 0 {
		headers["X-Nuon-API-Next"] = ""
		return []app.OtelLogRecord{}, headers, nil
	}

	var otelLogRecords []app.OtelLogRecord

	if order == "asc" {
		// ASC: Forward pagination - get records newer than cursor
		res := s.chDB.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where("log_stream_id = ?", logStreamID)

		if cursor > 0 {
			res = res.Where("toUnixTimestamp64Nano(timestamp) > ?", cursor)
		}

		res = applyLogFilters(res, filters).
			Order("timestamp ASC").
			Limit(PageSize).
			Find(&otelLogRecords)
		if res.Error != nil {
			return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs")
		}

		// Determine next cursor
		if len(otelLogRecords) < PageSize {
			headers["X-Nuon-API-Next"] = ""
		} else {
			last := otelLogRecords[len(otelLogRecords)-1]
			headers["X-Nuon-API-Next"] = fmt.Sprintf("%d", last.Timestamp.UnixNano())
		}

	} else {
		// DESC: Reverse pagination using ASC query + offset calculation
		// We use ASC ordering because ClickHouse is optimized for forward scans on time-series data
		var recordCount int64

		if cursor == 0 {
			// First page - use total count
			recordCount = totalCount
		} else {
			// Subsequent pages - count records strictly before cursor (exclusive)
			countQ := s.chDB.WithContext(ctx).
				Model(&app.OtelLogRecord{}).
				Where("org_id = ?", orgID).
				Where("log_stream_id = ?", logStreamID).
				Where("toUnixTimestamp64Nano(timestamp) < ?", cursor)
			countQ = applyLogFilters(countQ, filters)
			if res := countQ.Count(&recordCount); res.Error != nil {
				return nil, headers, errors.Wrap(res.Error, "unable to count remaining logs")
			}
		}

		// No more records
		if recordCount == 0 {
			headers["X-Nuon-API-Next"] = ""
			return []app.OtelLogRecord{}, headers, nil
		}

		// Calculate offset to get the last PageSize records from the available set
		offset := recordCount - int64(PageSize)
		if offset < 0 {
			offset = 0
		}

		// Query with ASC order, applying cursor filter and offset
		res := s.chDB.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where("log_stream_id = ?", logStreamID)

		if cursor > 0 {
			res = res.Where("toUnixTimestamp64Nano(timestamp) < ?", cursor)
		}

		res = applyLogFilters(res, filters).
			Order("timestamp ASC").
			Offset(int(offset)).
			Limit(PageSize).
			Find(&otelLogRecords)
		if res.Error != nil {
			return nil, headers, errors.Wrap(res.Error, "unable to retrieve logs")
		}

		// Reverse the results in memory to get DESC order
		for i, j := 0, len(otelLogRecords)-1; i < j; i, j = i+1, j-1 {
			otelLogRecords[i], otelLogRecords[j] = otelLogRecords[j], otelLogRecords[i]
		}

		// Determine next cursor
		// If offset was 0, we've retrieved all remaining records
		if offset == 0 {
			headers["X-Nuon-API-Next"] = ""
		} else {
			// Last element after reversal is the oldest timestamp in this batch
			last := otelLogRecords[len(otelLogRecords)-1]
			headers["X-Nuon-API-Next"] = fmt.Sprintf("%d", last.Timestamp.UnixNano())
		}
	}

	return otelLogRecords, headers, nil
}
