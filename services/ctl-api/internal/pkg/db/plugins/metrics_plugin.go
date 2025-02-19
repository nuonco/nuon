package plugins

import (
	"context"
	"reflect"
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type contextKey string

const (
	defaultContextKey contextKey = "gorm_metrics_plugin"
)

var _ gorm.Plugin = (*metricsWriterPlugin)(nil)

// This is a plugin that emits well-formed metrics to datadog based on queries/operations performed by gorm.
//
// It is semi-inspired by https://github.com/go-gorm/prometheus/blob/master/prometheus.go which takes this a step
// further by pulling in database metrics and emitting them via prometheus, however prometheus is lower level than what
// we have here.
func NewMetricsPlugin(mw metrics.Writer, dbType string) *metricsWriterPlugin {
	return &metricsWriterPlugin{
		metricsWriter: mw,
		dbType:        dbType,
	}
}

type metricsWriterPlugin struct {
	metricsWriter metrics.Writer
	dbType        string
}

func (m *metricsWriterPlugin) Name() string {
	return "metrics-writer"
}

func (m *metricsWriterPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Create().Before("*").Register("before_all", m.beforeAll)
	db.Callback().Create().After("*").Register("after_all", m.afterAll)

	return nil
}

func (m *metricsWriterPlugin) beforeAll(tx *gorm.DB) {
	ctx := tx.Statement.Context
	ts := time.Now()

	ctx = context.WithValue(ctx, defaultContextKey, ts)
	tx.Statement.Context = ctx

	metrics, err := cctx.MetricsContextFromGinContext(ctx)
	if err != nil {
		return
	}

	metrics.DBQueryCount += 1
}

func (m *metricsWriterPlugin) afterAll(tx *gorm.DB) {
	ctx := tx.Statement.Context
	schema := tx.Statement.Schema

	val := ctx.Value(defaultContextKey)
	if val == nil {
		return
	}
	startTS := val.(time.Time)

	tags := []string{}
	if schema != nil {
		tags = append(tags, "table:"+schema.Table)
		tags = append(tags, "db_type:"+m.dbType)
	}

	metricCtx, err := cctx.MetricsContextFromGinContext(ctx)
	if err != nil {
		return
	}

	tags = append(tags, []string{
		"endpoint:" + metricCtx.Endpoint,
		"context:" + metricCtx.Context,
		"method:" + metricCtx.Method,
		"org_id:" + metricCtx.OrgID,
	}...)

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

	m.metricsWriter.Incr("gorm_operation", tags)
	m.metricsWriter.Timing("gorm_operation_latency", time.Since(startTS), tags)
	m.metricsWriter.Gauge("gorm_operation.response_size", float64(respSize), tags)
	m.metricsWriter.Gauge("gorm_operation.preload_count", preloadCount, tags)
}
