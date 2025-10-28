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

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type contextKey string

const (
	defaultContextKey contextKey    = "gorm_metrics_plugin"
	targetLatency     time.Duration = time.Millisecond * 50
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
)

func (m *metricsWriterPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Create().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, CreateOperation) })
	db.Callback().Create().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, CreateOperation) })
	db.Callback().Query().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, QueryOperation) })
	db.Callback().Query().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, QueryOperation) })
	db.Callback().Update().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, UpdateOperation) })
	db.Callback().Update().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, UpdateOperation) })
	db.Callback().Delete().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, DeleteOperation) })
	db.Callback().Delete().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, DeleteOperation) })
	db.Callback().Raw().Before("*").Register("before_all", func(tx *gorm.DB) { m.beforeAll(tx, RawOperation) })
	db.Callback().Raw().After("*").Register("after_all", func(tx *gorm.DB) { m.afterAll(tx, RawOperation) })

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

func (m *metricsWriterPlugin) afterAll(tx *gorm.DB, operationType OperationType) {
	ctx := tx.Statement.Context
	schema := tx.Statement.Schema

	val := ctx.Value(defaultContextKey)
	if val == nil {
		return
	}
	startTS := val.(time.Time)
	dur := time.Since(startTS)
	withinTargetLatency := time.Since(startTS) < targetLatency

	tags := []string{}
	if schema != nil {
		tags = append(tags,
			"table:"+schema.Table,
			"db_type:"+m.dbType,
			"db_operation_type:"+string(operationType),
			"within_target_latency:"+strconv.FormatBool(withinTargetLatency),
		)
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
	m.metricsWriter.Gauge("gorm_operation.response_size", float64(respSize), tags)
	m.metricsWriter.Gauge("gorm_operation.preload_count", preloadCount, tags)
	m.metricsWriter.Gauge("gorm_operation.rows_affected", float64(tx.RowsAffected), tags)

	largeResultSetThreshold := 350

	if respSize >= largeResultSetThreshold {
		m.l.Warn(ctx, "large response_size",
			"table", schema.Table,
			"request_uri", metricCtx.RequestURI,
			"endpoint", metricCtx.Endpoint,
			"method", metricCtx.Method,
			"org_id", metricCtx.OrgID,
			"response_size", respSize,
		)
	}

	if m.dbType == "ch" {
		return
	}

	if dur < targetLatency {
		return
	}

	m.metricsWriter.Event(&statsd.Event{
		Title: "Slow query" + metricCtx.Endpoint,
		Text: fmt.Sprintf("Slow query identified for table %s and endpoint %s\n\nPrepared SQL: %s\nVars: %v\n",
			schema.Table,
			metricCtx.Endpoint,
			tx.Statement.SQL.String(),
			tx.Statement.Vars,
		),
		Tags: tags,
	})

	// Log slow queries
	m.l.Error(ctx, "Slow query identified",
		"table", schema.Table,
		"request_uri", metricCtx.RequestURI,
		"endpoint", metricCtx.Endpoint,
		"method", metricCtx.Method,
		"org_id", metricCtx.OrgID,
		"latency_ms", dur.Milliseconds(),
		"prepared_sql", tx.Statement.SQL.String(),
		"vars", tx.Statement.Vars,
		"final_sql", tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...),
	)
}
