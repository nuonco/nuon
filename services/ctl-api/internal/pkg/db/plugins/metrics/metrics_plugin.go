package metrics

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gorm.io/gorm"
	"moul.io/zapgorm2"

	"github.com/DataDog/datadog-go/v5/statsd"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

type contextKey string

const (
	defaultContextKey      contextKey    = "gorm_metrics_plugin"
	statementStartKey      contextKey    = "gorm_metrics_plugin_statement_start"
	statementDurationKey   contextKey    = "gorm_metrics_plugin_statement_duration"
	targetLatency          time.Duration = time.Millisecond * 500
	statementTargetLatency time.Duration = time.Millisecond * 200
	maxEventTextLen        int           = 4096
	eventTextTruncSuffix   string        = "... [truncated]"
)

var _ gorm.Plugin = (*metricsWriterPlugin)(nil)

// This is a plugin that emits well-formed metrics to datadog based on queries/operations performed by gorm.
//
// It is semi-inspired by https://github.com/go-gorm/prometheus/blob/master/prometheus.go which takes this a step
// further by pulling in database metrics and emitting them via prometheus, however prometheus is lower level than what
// we have here.
func NewMetricsPlugin(mw metrics.Writer, dbType string, L *zapgorm2.Logger) *metricsWriterPlugin {
	return &metricsWriterPlugin{
		metricsWriter: mw,
		dbType:        dbType,
		l:             L,
	}
}

type metricsWriterPlugin struct {
	metricsWriter metrics.Writer
	dbType        string
	l             *zapgorm2.Logger
}

func (m *metricsWriterPlugin) Name() string {
	return "metrics-writer"
}

type OperationType string

const (
	CreateOperation OperationType = "create"
	QueryOperation  OperationType = "query"
	UpdateOperation OperationType = "update"
	DeleteOperation OperationType = "delete"
	RawOperation    OperationType = "raw"
	RowOperation    OperationType = "row"
)

func (m *metricsWriterPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Create().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, CreateOperation) })
	db.Callback().Create().After("gorm:save_before_associations").Before("gorm:create").Register("before_statement", m.beforeStatement)
	db.Callback().Create().After("gorm:create").Before("gorm:save_after_associations").Register("after_statement", m.afterStatement)
	db.Callback().Create().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, CreateOperation) })

	db.Callback().Query().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, QueryOperation) })
	db.Callback().Query().Before("gorm:query").Register("before_statement", m.beforeStatement)
	db.Callback().Query().After("gorm:query").Before("gorm:preload").Register("after_statement", m.afterStatement)
	db.Callback().Query().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, QueryOperation) })

	db.Callback().Update().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, UpdateOperation) })
	db.Callback().Update().After("gorm:save_before_associations").Before("gorm:update").Register("before_statement", m.beforeStatement)
	db.Callback().Update().After("gorm:update").Before("gorm:save_after_associations").Register("after_statement", m.afterStatement)
	db.Callback().Update().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, UpdateOperation) })

	db.Callback().Delete().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, DeleteOperation) })
	db.Callback().Delete().After("gorm:delete_before_associations").Before("gorm:delete").Register("before_statement", m.beforeStatement)
	db.Callback().Delete().After("gorm:delete").Register("after_statement", m.afterStatement)
	db.Callback().Delete().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, DeleteOperation) })

	db.Callback().Raw().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, RawOperation) })
	db.Callback().Raw().Before("gorm:raw").Register("before_statement", m.beforeStatement)
	db.Callback().Raw().After("gorm:raw").Register("after_statement", m.afterStatement)
	db.Callback().Raw().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, RawOperation) })

	db.Callback().Row().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, RowOperation) })
	db.Callback().Row().Before("gorm:row").Register("before_statement", m.beforeStatement)
	db.Callback().Row().After("gorm:row").Register("after_statement", m.afterStatement)
	db.Callback().Row().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, RowOperation) })

	return nil
}

func (m *metricsWriterPlugin) beforeAll(tx *gorm.DB, operationType OperationType) {
	ctx := tx.Statement.Context
	ts := time.Now()

	ctx = context.WithValue(ctx, defaultContextKey, ts)
	tx.Statement.Context = ctx

	metrics, err := cctx.MetricsContextFromGinContext(ctx)
	if err != nil {
		return
	}

	metrics.DBType = m.dbType
	metrics.DBOperationType = string(operationType)
	metrics.DBQueryCount += 1
}

// beforeStatement/afterStatement bracket the core SQL statement (the SELECT/INSERT/UPDATE/DELETE itself), excluding
// the preload and save/delete-association phases. This isolates the single contained psql request so afterAll can
// report gorm_statement_latency per statement rather than cumulating association/preload round-trips.
func (m *metricsWriterPlugin) beforeStatement(tx *gorm.DB) {
	tx.Statement.Context = context.WithValue(tx.Statement.Context, statementStartKey, time.Now())
}

func (m *metricsWriterPlugin) afterStatement(tx *gorm.DB) {
	ctx := tx.Statement.Context
	val := ctx.Value(statementStartKey)
	if val == nil {
		return
	}
	tx.Statement.Context = context.WithValue(ctx, statementDurationKey, time.Since(val.(time.Time)))
}

func (m *metricsWriterPlugin) afterAll(tx *gorm.DB, operationType OperationType) {
	ctx := tx.Statement.Context
	schema := tx.Statement.Schema

	tableName := "raw_sql"
	if schema != nil {
		tableName = schema.Table
	}

	val := ctx.Value(defaultContextKey)
	if val == nil {
		return
	}
	startTS := val.(time.Time)
	dur := time.Since(startTS)
	withinTargetLatency := time.Since(startTS) < targetLatency

	// statement latency isolates the root query (no preloads); falls back to dur for ops with no preload phase.
	stmtDur := dur
	if v := ctx.Value(statementDurationKey); v != nil {
		stmtDur = v.(time.Duration)
	}

	tags := []string{
		"table:" + tableName,
		"db_type:" + m.dbType,
		"db_operation_type:" + string(operationType),
		"within_target_latency:" + strconv.FormatBool(withinTargetLatency),
	}
	if m.dbType == "psql" {
		tags = append(tags, "pool:"+string(routing.DecisionFromContext(ctx)))
	}

	metricCtx, err := cctx.MetricsContextFromGinContext(ctx)
	if err != nil {
		return
	}

	tags = append(tags,
		"context:"+metricCtx.Context,
		"method:"+metricCtx.Method,
		"endpoint:"+metricCtx.Endpoint,
		"org_id:"+metricCtx.OrgID,
		"namespace:"+metricCtx.Namespace,
	)

	respSize := 0
	if tx.Statement.ReflectValue.IsValid() {
		if tx.Statement.ReflectValue.Kind() == reflect.Slice {
			respSize = tx.Statement.ReflectValue.Len()
		} else {
			if !tx.Statement.ReflectValue.IsZero() {
				respSize = 1
			}
		}
	}

	preloadCount := float64(len(tx.Statement.Preloads))

	m.metricsWriter.Timing("gorm_operation_latency", dur, tags)
	m.metricsWriter.Timing("gorm_operation.statement_latency", stmtDur, tags)
	m.metricsWriter.Gauge("gorm_operation.response_size", float64(respSize), tags)
	m.metricsWriter.Gauge("gorm_operation.preload_count", preloadCount, tags)
	m.metricsWriter.Gauge("gorm_operation.rows_affected", float64(tx.RowsAffected), tags)

	largeResultSetThreshold := 350

	if respSize >= largeResultSetThreshold {
		m.l.Warn(ctx, "large response_size",
			"table", tableName,
			"request_uri", metricCtx.RequestURI,
			"endpoint", metricCtx.Endpoint,
			"method", metricCtx.Method,
			"org_id", metricCtx.OrgID,
			"namespace", metricCtx.Namespace,
			"response_size", respSize,
		)
	}

	if m.dbType == "ch" {
		return
	}

	// statement-level slow query: the root SQL itself exceeded target, excluding preloads.
	if stmtDur >= statementTargetLatency {
		stmtEventText := fmt.Sprintf("Slow gorm statement identified for table %s and endpoint %s (latency: %dms)\n\nPrepared SQL: %s\nVars: %v\n",
			tableName,
			metricCtx.Endpoint,
			stmtDur.Milliseconds(),
			tx.Statement.SQL.String(),
			tx.Statement.Vars,
		)
		if len(stmtEventText) > maxEventTextLen {
			stmtEventText = stmtEventText[:maxEventTextLen-len(eventTextTruncSuffix)] + eventTextTruncSuffix
		}

		m.metricsWriter.Event(&statsd.Event{
			Title: fmt.Sprintf("Slow gorm statement: %s (%dms)", metricCtx.Endpoint, stmtDur.Milliseconds()),
			Text:  stmtEventText,
			Tags:  tags,
		})
	}

	if dur < targetLatency {
		return
	}

	eventText := fmt.Sprintf("Slow query identified for table %s and endpoint %s (latency: %dms)\n\nPrepared SQL: %s\nVars: %v\n",
		tableName,
		metricCtx.Endpoint,
		dur.Milliseconds(),
		tx.Statement.SQL.String(),
		tx.Statement.Vars,
	)
	if len(eventText) > maxEventTextLen {
		eventText = eventText[:maxEventTextLen-len(eventTextTruncSuffix)] + eventTextTruncSuffix
	}

	m.metricsWriter.Event(&statsd.Event{
		Title: fmt.Sprintf("Slow query: %s (%dms)", metricCtx.Endpoint, dur.Milliseconds()),
		Text:  eventText,
		Tags:  tags,
	})

	// Log slow queries
	m.l.Error(ctx, "Slow query identified",
		"table", tableName,
		"request_uri", metricCtx.RequestURI,
		"endpoint", metricCtx.Endpoint,
		"method", metricCtx.Method,
		"org_id", metricCtx.OrgID,
		"namespace", metricCtx.Namespace,
		"latency_ms", dur.Milliseconds(),
		"prepared_sql", tx.Statement.SQL.String(),
		"vars", tx.Statement.Vars,
		"final_sql", tx.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...),
	)
}
