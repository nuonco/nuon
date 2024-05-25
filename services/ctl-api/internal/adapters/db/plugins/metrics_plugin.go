package plugins

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	metrics_middleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/metrics"
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
func NewMetricsPlugin(mw metrics.Writer) *metricsWriterPlugin {
	return &metricsWriterPlugin{
		metricsWriter: mw,
	}
}

type metricsWriterPlugin struct {
	metricsWriter metrics.Writer
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
	}

	metricCtx, err := metrics_middleware.FromContext(ctx)
	if err == nil {
		tags = append(tags, []string{
			"endpoint:" + metricCtx.Endpoint,
			"context:" + metricCtx.Context,
			"method:" + metricCtx.Method,
			"org_id:" + metricCtx.OrgID,
		}...)
	}

	m.metricsWriter.Incr("gorm_operation", tags)
	m.metricsWriter.Timing("gorm_operation_latency", time.Since(startTS), tags)
}
