package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

	// maxAttrFilters caps the total number of generic attribute filters
	// (attr / resource_attr / scope_attr combined) accepted on a single
	// request. Each filter adds two predicates (mapContains + bracket
	// access) — bounding the count keeps query plans sane.
	maxAttrFilters int = 16
)

var attrKeyRe = regexp.MustCompile(nestedAttributeRegex)

// kvFilter represents a single map-attribute predicate (key=value).
type kvFilter struct {
	key string
	val string
}

// logFilters holds optional filter values parsed from query parameters.
//
// The endpoint exposes the full OTEL log data model:
//
//   - Top-level CH columns (Timestamp, SeverityText/Number, ServiceName,
//     ScopeName/Version, ResourceSchemaURL, ScopeSchemaURL, Trace/SpanID,
//     Body) — filtered with direct SQL.
//   - Internal scoping columns (RunnerID, RunnerJobID, RunnerGroupID,
//     RunnerJobExecutionID, RunnerJobExecutionStep) — most live on the
//     primary key/order key, so they prune efficiently.
//   - Map columns (LogAttributes, ResourceAttributes, ScopeAttributes) —
//     use mapContains() AND bracket equality so the bloom_filter skip
//     indexes on mapKeys/mapValues (migration 06) can be leveraged.
//   - Typed shortcuts for common LogAttributes (nuon.tool, helm.*, tf.*,
//     k8s.*) for ergonomics — these are sugar over the generic attr= path.
type logFilters struct {
	// Time range — push-down on Timestamp prunes by the toDate(timestamp)
	// PARTITION BY clause (see otel_log_record.go GetTableOptions()).
	startTime time.Time
	endTime   time.Time

	// Top-level OTEL columns
	serviceNames       []string
	scopeNames         []string
	scopeVersions      []string
	severityTexts      []string
	severityNumberMin  int
	severityNumberMax  int
	resourceSchemaURLs []string
	scopeSchemaURLs    []string
	traceID            string
	spanID             string
	traceFlags         int
	traceFlagsSet      bool

	// Internal scoping (primary-key / order-key columns where applicable)
	runnerID               string
	runnerJobID            string
	runnerGroupID          string
	runnerJobExecutionID   string
	runnerJobExecutionStep string

	// Typed log_attributes shortcuts
	tools           []string
	helmReleaseName string
	helmChartName   string
	helmChartID     string
	helmNamespace   string
	helmOperation   string
	tfWorkspaceID   string
	tfOperation     string
	k8sKind         string
	k8sNamespace    string
	k8sName         string
	k8sOperation    string

	// Generic attribute filters (escape hatch for the long tail)
	logAttrs      []kvFilter
	resourceAttrs []kvFilter
	scopeAttrs    []kvFilter

	// Body substring (case-insensitive)
	bodyContains string
}

func parseLogFilters(ctx *gin.Context) (logFilters, error) {
	f := logFilters{
		serviceNames:       ctx.QueryArray("service_name"),
		scopeNames:         ctx.QueryArray("scope_name"),
		scopeVersions:      ctx.QueryArray("scope_version"),
		severityTexts:      ctx.QueryArray("severity_text"),
		resourceSchemaURLs: ctx.QueryArray("resource_schema_url"),
		scopeSchemaURLs:    ctx.QueryArray("scope_schema_url"),

		tools:           ctx.QueryArray("tool"),
		helmReleaseName: ctx.Query("helm_release_name"),
		helmChartName:   ctx.Query("helm_chart_name"),
		helmChartID:     ctx.Query("helm_chart_id"),
		helmNamespace:   ctx.Query("helm_namespace"),
		helmOperation:   ctx.Query("helm_operation"),
		tfWorkspaceID:   ctx.Query("tf_workspace_id"),
		tfOperation:     ctx.Query("tf_operation"),
		k8sKind:         ctx.Query("k8s_kind"),
		k8sNamespace:    ctx.Query("k8s_namespace"),
		k8sName:         ctx.Query("k8s_name"),
		k8sOperation:    ctx.Query("k8s_operation"),

		runnerID:               ctx.Query("runner_id"),
		runnerJobID:            ctx.Query("runner_job_id"),
		runnerGroupID:          ctx.Query("runner_group_id"),
		runnerJobExecutionID:   ctx.Query("runner_job_execution_id"),
		runnerJobExecutionStep: ctx.Query("runner_job_execution_step"),

		traceID: ctx.Query("trace_id"),
		spanID:  ctx.Query("span_id"),

		bodyContains: firstNonEmpty(ctx.Query("q"), ctx.Query("body_contains")),
	}

	// Time range — RFC3339.
	if v := ctx.Query("start_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, fmt.Errorf("invalid start_time, expected RFC3339: %w", err)
		}
		f.startTime = t
	}
	if v := ctx.Query("end_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, fmt.Errorf("invalid end_time, expected RFC3339: %w", err)
		}
		f.endTime = t
	}
	if !f.startTime.IsZero() && !f.endTime.IsZero() && f.endTime.Before(f.startTime) {
		return f, errors.New("end_time must be >= start_time")
	}

	// Severity number range (OTEL: TRACE=1..FATAL=24, stored as UInt8).
	if v := ctx.Query("severity_number_min"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 255 {
			return f, errors.New("invalid severity_number_min (expected 0-255)")
		}
		f.severityNumberMin = n
	}
	if v := ctx.Query("severity_number_max"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 255 {
			return f, errors.New("invalid severity_number_max (expected 0-255)")
		}
		f.severityNumberMax = n
	}

	// Trace flags (UInt8).
	if v := ctx.Query("trace_flags"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 255 {
			return f, errors.New("invalid trace_flags (expected 0-255)")
		}
		f.traceFlags = n
		f.traceFlagsSet = true
	}

	// Generic key:value attribute filters — repeatable across three columns.
	var err error
	if f.logAttrs, err = parseKVFilters(ctx.QueryArray("attr")); err != nil {
		return f, fmt.Errorf("invalid attr: %w", err)
	}
	if f.resourceAttrs, err = parseKVFilters(ctx.QueryArray("resource_attr")); err != nil {
		return f, fmt.Errorf("invalid resource_attr: %w", err)
	}
	if f.scopeAttrs, err = parseKVFilters(ctx.QueryArray("scope_attr")); err != nil {
		return f, fmt.Errorf("invalid scope_attr: %w", err)
	}
	if len(f.logAttrs)+len(f.resourceAttrs)+len(f.scopeAttrs) > maxAttrFilters {
		return f, fmt.Errorf("too many attribute filters (max %d)", maxAttrFilters)
	}

	return f, nil
}

// parseKVFilters parses repeatable key:value query params. The key is
// validated against nestedAttributeRegex; the value is taken verbatim
// (anything after the first ':'). Empty entries are skipped.
func parseKVFilters(raws []string) ([]kvFilter, error) {
	out := make([]kvFilter, 0, len(raws))
	for _, raw := range raws {
		if raw == "" {
			continue
		}
		idx := strings.IndexByte(raw, ':')
		if idx <= 0 || idx == len(raw)-1 {
			return nil, fmt.Errorf("expected key:value, got %q", raw)
		}
		key, val := raw[:idx], raw[idx+1:]
		if !attrKeyRe.MatchString(key) {
			return nil, fmt.Errorf("invalid key %q", key)
		}
		out = append(out, kvFilter{key: key, val: val})
	}
	return out, nil
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
	// Time range — partition pruning on toDate(timestamp_time).
	if !f.startTime.IsZero() {
		db = db.Where("timestamp >= ?", f.startTime)
	}
	if !f.endTime.IsZero() {
		db = db.Where("timestamp <= ?", f.endTime)
	}

	// Top-level columns.
	if len(f.serviceNames) > 0 {
		db = db.Where("service_name IN ?", f.serviceNames)
	}
	if len(f.scopeNames) > 0 {
		db = db.Where("scope_name IN ?", f.scopeNames)
	}
	if len(f.scopeVersions) > 0 {
		db = db.Where("scope_version IN ?", f.scopeVersions)
	}
	if len(f.severityTexts) > 0 {
		db = db.Where("severity_text IN ?", f.severityTexts)
	}
	if f.severityNumberMin > 0 {
		db = db.Where("severity_number >= ?", f.severityNumberMin)
	}
	if f.severityNumberMax > 0 {
		db = db.Where("severity_number <= ?", f.severityNumberMax)
	}
	if len(f.resourceSchemaURLs) > 0 {
		db = db.Where("resource_schema_url IN ?", f.resourceSchemaURLs)
	}
	if len(f.scopeSchemaURLs) > 0 {
		db = db.Where("scope_schema_url IN ?", f.scopeSchemaURLs)
	}

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
	if f.traceFlagsSet {
		db = db.Where("trace_flags = ?", f.traceFlags)
	}

	// Internal scoping. runner_job_id is part of the ORDER BY tuple — passing
	// it lets ClickHouse skip whole granules. Step pages already know this ID.
	if f.runnerID != "" {
		db = db.Where("runner_id = ?", f.runnerID)
	}
	if f.runnerJobID != "" {
		db = db.Where("runner_job_id = ?", f.runnerJobID)
	}
	if f.runnerGroupID != "" {
		db = db.Where("runner_group_id = ?", f.runnerGroupID)
	}
	if f.runnerJobExecutionID != "" {
		db = db.Where("runner_job_execution_id = ?", f.runnerJobExecutionID)
	}
	if f.runnerJobExecutionStep != "" {
		db = db.Where("runner_job_execution_step = ?", f.runnerJobExecutionStep)
	}

	// Typed log_attributes shortcuts. mapContains() exercises the
	// mapKeys bloom_filter; the bracket equality exercises the mapValues
	// bloom_filter (see migration 06). AND'ing both lets ClickHouse prune
	// granules from either side.
	addAttrEq := func(d *gorm.DB, key, val string) *gorm.DB {
		if val == "" {
			return d
		}
		return d.Where("mapContains(log_attributes, ?) AND log_attributes[?] = ?", key, key, val)
	}
	addAttrIn := func(d *gorm.DB, key string, vals []string) *gorm.DB {
		if len(vals) == 0 {
			return d
		}
		return d.Where("mapContains(log_attributes, ?) AND log_attributes[?] IN ?", key, key, vals)
	}

	db = addAttrIn(db, "nuon.tool", f.tools)
	db = addAttrEq(db, "helm.release_name", f.helmReleaseName)
	db = addAttrEq(db, "helm.chart_name", f.helmChartName)
	db = addAttrEq(db, "helm.chart_id", f.helmChartID)
	db = addAttrEq(db, "helm.namespace", f.helmNamespace)
	db = addAttrEq(db, "helm.operation", f.helmOperation)
	db = addAttrEq(db, "tf.workspace_id", f.tfWorkspaceID)
	db = addAttrEq(db, "tf.operation", f.tfOperation)
	db = addAttrEq(db, "k8s.kind", f.k8sKind)
	db = addAttrEq(db, "k8s.namespace", f.k8sNamespace)
	db = addAttrEq(db, "k8s.name", f.k8sName)
	db = addAttrEq(db, "k8s.operation", f.k8sOperation)

	// Generic key:value filters across the three OTEL attribute maps.
	db = applyMapKVs(db, "log_attributes", f.logAttrs)
	db = applyMapKVs(db, "resource_attributes", f.resourceAttrs)
	db = applyMapKVs(db, "scope_attributes", f.scopeAttrs)

	if f.bodyContains != "" {
		db = db.Where("body ILIKE ?", "%"+f.bodyContains+"%")
	}
	return db
}

// applyMapKVs is the generic counterpart to addAttrEq: it pushes one
// (mapContains AND bracket =) pair per filter on the named Map column.
// `col` is restricted to the three OTEL attribute columns by the caller.
func applyMapKVs(db *gorm.DB, col string, kvs []kvFilter) *gorm.DB {
	if len(kvs) == 0 {
		return db
	}
	clause := fmt.Sprintf("mapContains(%s, ?) AND %s[?] = ?", col, col)
	for _, kv := range kvs {
		db = db.Where(clause, kv.key, kv.key, kv.val)
	}
	return db
}

// @ID						LogStreamReadLogs
// @Summary				read a log stream's logs
// @Description.markdown	log_stream_read_logs.md
// @Param					log_stream_id		path	string	true	"log stream ID"
// @Param					X-Nuon-API-Offset	header	string	false	"log stream offset"
// @Param					order				query	string	false	"sort direction"	default(asc)
// @Param					start_time			query	string		false	"only return records with timestamp >= start_time (RFC3339)"
// @Param					end_time			query	string		false	"only return records with timestamp <= end_time (RFC3339)"
// @Param					service_name		query	[]string	false	"filter by service_name (repeatable)"
// @Param					scope_name			query	[]string	false	"filter by scope_name (repeatable; e.g. oteljob, system)"
// @Param					scope_version		query	[]string	false	"filter by scope_version (repeatable)"
// @Param					resource_schema_url	query	[]string	false	"filter by resource_schema_url (repeatable)"
// @Param					scope_schema_url	query	[]string	false	"filter by scope_schema_url (repeatable)"
// @Param					severity_text		query	[]string	false	"filter by severity_text (repeatable; INFO/WARN/ERROR/...)"
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
// @Param					tool				query	[]string	false	"filter by log_attributes['nuon.tool'] (repeatable; e.g. helm, terraform, kubernetes_manifest, runner)"
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
// @Param					attr				query	[]string	false	"generic log_attributes filter as 'key:value' (repeatable, max 16 across all attr params)"
// @Param					resource_attr		query	[]string	false	"generic resource_attributes filter as 'key:value' (repeatable, max 16 across all attr params)"
// @Param					scope_attr			query	[]string	false	"generic scope_attributes filter as 'key:value' (repeatable, max 16 across all attr params)"
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

	filters, err := parseLogFilters(ctx)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	logs, headers, readErr := s.getLogStreamLogs(ctx, logStreamID, orgID, cursor, order, filters)
	if readErr != nil {
		ctx.Error(errors.Wrap(readErr, "unable to read runner logs"))
		return
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
