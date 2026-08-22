package tracing

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

type contextKey string

const (
	spanContextKey          contextKey = "gorm_tracing_plugin_span"
	statementSpanContextKey contextKey = "gorm_tracing_plugin_statement_span"
)

var _ gorm.Plugin = (*tracingPlugin)(nil)

// This plugin emits OpenTelemetry spans for every gorm operation, mirroring the tags emitted by the
// metrics plugin so high-cardinality dimensions (endpoint, table, org) live on traces instead of
// metric tags. Preload queries re-enter the Query callback pipeline with the parent operation's
// context, so they show up as child spans automatically.
//
// Each operation also gets a child statement span bracketing only the core SQL, using the same
// callback positions as the metrics plugin so the span duration lines up with
// gorm_operation.statement_latency rather than the preload-inclusive gorm_operation_latency.
//
// Spans only ever carry the prepared SQL with placeholders, never bound values.
func NewTracingPlugin(dbType string) *tracingPlugin {
	return &tracingPlugin{
		dbType: dbType,
		tracer: otel.Tracer("github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/tracing"),
	}
}

type tracingPlugin struct {
	dbType string
	tracer trace.Tracer
}

func (p *tracingPlugin) Name() string {
	return "otel-tracing"
}

type operationType string

const (
	createOperation operationType = "create"
	queryOperation  operationType = "query"
	updateOperation operationType = "update"
	deleteOperation operationType = "delete"
	rawOperation    operationType = "raw"
	rowOperation    operationType = "row"
)

func (p *tracingPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Create().Before("*").Register("otel_before", func(tx *gorm.DB) { p.before(tx, createOperation) })
	db.Callback().Create().After("gorm:save_before_associations").Before("gorm:create").Register("otel_before_statement", func(tx *gorm.DB) { p.beforeStatement(tx, createOperation) })
	db.Callback().Create().After("gorm:create").Before("gorm:save_after_associations").Register("otel_after_statement", func(tx *gorm.DB) { p.afterStatement(tx, createOperation) })
	db.Callback().Create().After("*").Register("otel_after", func(tx *gorm.DB) { p.after(tx, createOperation) })

	db.Callback().Query().Before("*").Register("otel_before", func(tx *gorm.DB) { p.before(tx, queryOperation) })
	db.Callback().Query().Before("gorm:query").Register("otel_before_statement", func(tx *gorm.DB) { p.beforeStatement(tx, queryOperation) })
	db.Callback().Query().After("gorm:query").Before("gorm:preload").Register("otel_after_statement", func(tx *gorm.DB) { p.afterStatement(tx, queryOperation) })
	db.Callback().Query().After("*").Register("otel_after", func(tx *gorm.DB) { p.after(tx, queryOperation) })

	db.Callback().Update().Before("*").Register("otel_before", func(tx *gorm.DB) { p.before(tx, updateOperation) })
	db.Callback().Update().After("gorm:save_before_associations").Before("gorm:update").Register("otel_before_statement", func(tx *gorm.DB) { p.beforeStatement(tx, updateOperation) })
	db.Callback().Update().After("gorm:update").Before("gorm:save_after_associations").Register("otel_after_statement", func(tx *gorm.DB) { p.afterStatement(tx, updateOperation) })
	db.Callback().Update().After("*").Register("otel_after", func(tx *gorm.DB) { p.after(tx, updateOperation) })

	db.Callback().Delete().Before("*").Register("otel_before", func(tx *gorm.DB) { p.before(tx, deleteOperation) })
	db.Callback().Delete().After("gorm:delete_before_associations").Before("gorm:delete").Register("otel_before_statement", func(tx *gorm.DB) { p.beforeStatement(tx, deleteOperation) })
	db.Callback().Delete().After("gorm:delete").Register("otel_after_statement", func(tx *gorm.DB) { p.afterStatement(tx, deleteOperation) })
	db.Callback().Delete().After("*").Register("otel_after", func(tx *gorm.DB) { p.after(tx, deleteOperation) })

	db.Callback().Raw().Before("*").Register("otel_before", func(tx *gorm.DB) { p.before(tx, rawOperation) })
	db.Callback().Raw().Before("gorm:raw").Register("otel_before_statement", func(tx *gorm.DB) { p.beforeStatement(tx, rawOperation) })
	db.Callback().Raw().After("gorm:raw").Register("otel_after_statement", func(tx *gorm.DB) { p.afterStatement(tx, rawOperation) })
	db.Callback().Raw().After("*").Register("otel_after", func(tx *gorm.DB) { p.after(tx, rawOperation) })

	db.Callback().Row().Before("*").Register("otel_before", func(tx *gorm.DB) { p.before(tx, rowOperation) })
	db.Callback().Row().Before("gorm:row").Register("otel_before_statement", func(tx *gorm.DB) { p.beforeStatement(tx, rowOperation) })
	db.Callback().Row().After("gorm:row").Register("otel_after_statement", func(tx *gorm.DB) { p.afterStatement(tx, rowOperation) })
	db.Callback().Row().After("*").Register("otel_after", func(tx *gorm.DB) { p.after(tx, rowOperation) })

	return nil
}

func (p *tracingPlugin) before(tx *gorm.DB, op operationType) {
	spanParent := tx.Statement.Context
	// Handlers pass the *gin.Context to WithContext, and other callbacks (e.g. routing) may
	// have wrapped it by now. Gin only exposes request-context values (like the root HTTP
	// span) through Value when engine.ContextWithFallback is on, which it isn't, so recover
	// the gin context through its self-key and parent from the request context, while keeping
	// the original statement chain for Keys-based values like MetricContext.
	if gc, ok := spanParent.Value(gin.ContextKey).(*gin.Context); ok && gc.Request != nil {
		spanParent = gc.Request.Context()
	}

	_, span := p.tracer.Start(spanParent, "gorm."+string(op), trace.WithSpanKind(trace.SpanKindClient))
	if !span.IsRecording() {
		span.End()
		return
	}
	ctx := trace.ContextWithSpan(tx.Statement.Context, span)
	ctx = context.WithValue(ctx, spanContextKey, span)
	tx.Statement.Context = ctx
}

// beforeStatement/afterStatement bracket the core SQL only, so the child span excludes the preload
// and association phases the parent span covers. The statement span is deliberately not pushed onto
// the trace context: preload sub-queries must keep parenting under the operation span.
func (p *tracingPlugin) beforeStatement(tx *gorm.DB, op operationType) {
	ctx := tx.Statement.Context
	if ctx.Value(spanContextKey) == nil {
		return
	}

	_, span := p.tracer.Start(ctx, "gorm."+string(op)+".statement", trace.WithSpanKind(trace.SpanKindClient))
	if !span.IsRecording() {
		span.End()
		return
	}
	tx.Statement.Context = context.WithValue(ctx, statementSpanContextKey, span)
}

func (p *tracingPlugin) afterStatement(tx *gorm.DB, op operationType) {
	val := tx.Statement.Context.Value(statementSpanContextKey)
	if val == nil {
		return
	}
	span := val.(trace.Span)
	defer span.End()

	if !span.IsRecording() {
		return
	}

	dbSystem := p.dbSystem()
	table := tableName(tx)

	// Distinct operation name from the parent span, otherwise the statement spans fold into the
	// same Datadog APM stats family and double-count trace.<operation>.* for every query.
	span.SetAttributes(
		attribute.String("operation.name", dbSystem+".statement"),
		attribute.String("resource.name", fmt.Sprintf("%s %s", op, table)),
		attribute.String("db.system", dbSystem),
		attribute.String("db.operation", string(op)),
		attribute.String("db.sql.table", table),
		attribute.String("db.statement", tx.Statement.SQL.String()),
	)

	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, tx.Error.Error())
	}
}

func (p *tracingPlugin) after(tx *gorm.DB, op operationType) {
	ctx := tx.Statement.Context

	val := ctx.Value(spanContextKey)
	if val == nil {
		return
	}
	span := val.(trace.Span)
	defer span.End()

	if !span.IsRecording() {
		return
	}

	table := tableName(tx)
	dbSystem := p.dbSystem()

	// The ddotel bridge turns the span name into the Datadog operation name, and Datadog
	// creates one APM stats metric family per operation name (trace.<operation>.*). Keep the
	// operation low-cardinality (postgresql.query / clickhouse.query, matching Datadog's own
	// database integrations) and put the op + table in resource.name, which becomes the
	// resource_name tag on those metrics.
	attrs := []attribute.KeyValue{
		attribute.String("operation.name", dbSystem+".query"),
		attribute.String("resource.name", fmt.Sprintf("%s %s", op, table)),
		attribute.String("db.system", dbSystem),
		attribute.String("db.operation", string(op)),
		attribute.String("db.sql.table", table),
		attribute.String("db.statement", tx.Statement.SQL.String()),
		attribute.Int64("db.rows_affected", tx.RowsAffected),
		attribute.Int("nuon.db.preload_count", len(tx.Statement.Preloads)),
		attribute.Int("nuon.db.response_size", responseSize(tx)),
	}
	if p.dbType == "psql" {
		attrs = append(attrs, attribute.String("nuon.db.pool", string(routing.DecisionFromContext(ctx))))
	}

	if metricCtx, err := cctx.MetricsContextFromGinContext(ctx); err == nil {
		attrs = append(attrs,
			attribute.String("nuon.context", metricCtx.Context),
			attribute.String("http.method", metricCtx.Method),
			attribute.String("http.route", metricCtx.Endpoint),
			attribute.String("nuon.org_id", metricCtx.OrgID),
			attribute.String("nuon.namespace", metricCtx.Namespace),
		)
	}

	span.SetAttributes(attrs...)

	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, tx.Error.Error())
	}
}

func (p *tracingPlugin) dbSystem() string {
	if p.dbType == "ch" {
		return "clickhouse"
	}
	return "postgresql"
}

func tableName(tx *gorm.DB) string {
	if tx.Statement.Schema != nil {
		return tx.Statement.Schema.Table
	}
	return "raw_sql"
}

func responseSize(tx *gorm.DB) int {
	if !tx.Statement.ReflectValue.IsValid() {
		return 0
	}
	if tx.Statement.ReflectValue.Kind() == reflect.Slice {
		return tx.Statement.ReflectValue.Len()
	}
	if !tx.Statement.ReflectValue.IsZero() {
		return 1
	}
	return 0
}
