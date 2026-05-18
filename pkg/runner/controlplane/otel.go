package controlplane

import (
	"context"
	"sync"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const controlPlaneRunnerID = "control-plane"

type OTELLogRecord struct {
	RunnerID               string
	RunnerGroupID          string
	RunnerJobID            string
	RunnerJobExecutionID   string
	RunnerJobExecutionStep string
	ResourceAttributes     map[string]string
	ResourceSchemaURL      string
	ScopeSchemaURL         string
	ScopeName              string
	ScopeVersion           string
	ScopeAttributes        map[string]string
	Timestamp              time.Time
	ServiceName            string
	SeverityNumber         int
	SeverityText           string
	Body                   string
	TraceID                string
	SpanID                 string
	TraceFlags             int
	LogAttributes          map[string]string
}

type OTELTraceRecord struct {
	RunnerID               string
	RunnerGroupID          string
	RunnerJobID            string
	RunnerJobExecutionID   string
	RunnerJobExecutionStep string
	Timestamp              time.Time
	ResourceAttributes     map[string]string
	ResourceSchemaURL      string
	ScopeSchemaURL         string
	ScopeName              string
	ScopeVersion           string
	ScopeAttributes        map[string]string
	TraceID                string
	SpanID                 string
	ParentSpanID           string
	TraceState             string
	SpanName               string
	SpanKind               string
	ServiceName            string
	SpanAttributes         map[string]string
	Duration               int64
	StatusCode             string
	StatusMessage          string
	EventsTimestamp        []time.Time
	EventsName             []string
	EventsAttributes       []map[string]string
	LinksTraceID           []string
	LinksSpanID            []string
	LinksState             []string
	LinksAttributes        []map[string]string
}

func (e *Executor) telemetryProviders(job *models.AppRunnerJob) (*sdklog.LoggerProvider, *sdktrace.TracerProvider) {
	if job.LogStreamID == "" {
		return nil, nil
	}
	rsrc := telemetryResource(job)
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(rsrc),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(&logExporter{client: e.client, logStreamID: job.LogStreamID},
				sdklog.WithExportMaxBatchSize(25),
				sdklog.WithExportInterval(time.Second),
			),
		),
	)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(rsrc),
		sdktrace.WithBatcher(&traceExporter{client: e.client, runnerID: runnerIDForJob(job)},
			sdktrace.WithMaxExportBatchSize(64),
			sdktrace.WithBatchTimeout(time.Second),
		),
	)
	return logProvider, traceProvider
}

func telemetryResource(job *models.AppRunnerJob) *resource.Resource {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", "runner"),
		attribute.String("runner.id", runnerIDForJob(job)),
		attribute.String("log_stream.id", job.LogStreamID),
	}
	if job.OrgID != "" {
		attrs = append(attrs, attribute.String("org.id", job.OrgID))
	}
	return resource.NewSchemaless(attrs...)
}

func runnerIDForJob(job *models.AppRunnerJob) string {
	if job.RunnerID != "" {
		return job.RunnerID
	}
	return controlPlaneRunnerID
}

func newJobLogger(lp *sdklog.LoggerProvider, sysLog *zap.Logger) *zap.Logger {
	zapCore := otelzap.NewCore("oteljob", otelzap.WithLoggerProvider(lp))
	if sysLog == nil {
		return zap.New(zapCore)
	}
	return zap.New(zapcore.NewTee(sysLog.Core(), zapCore))
}

func flushTelemetry(lp *sdklog.LoggerProvider, tp *sdktrace.TracerProvider, l *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if lp != nil {
		if err := lp.ForceFlush(ctx); err != nil {
			l.Warn("unable to flush control-plane job logs", zap.Error(err))
		}
		if err := lp.Shutdown(ctx); err != nil {
			l.Warn("unable to shutdown control-plane job log provider", zap.Error(err))
		}
	}
	if tp != nil {
		if err := tp.ForceFlush(ctx); err != nil {
			l.Warn("unable to flush control-plane job traces", zap.Error(err))
		}
		if err := tp.Shutdown(ctx); err != nil {
			l.Warn("unable to shutdown control-plane job trace provider", zap.Error(err))
		}
	}
}

type logExporter struct {
	client      Client
	logStreamID string

	mu     sync.Mutex
	closed bool
}

func (e *logExporter) Export(ctx context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	closed := e.closed
	e.mu.Unlock()
	if closed || len(records) == 0 {
		return nil
	}

	out := make([]OTELLogRecord, 0, len(records))
	for _, record := range records {
		resourceAttrs := attributeMap(record.Resource().Attributes())
		scope := record.InstrumentationScope()
		scopeAttrs := attributeMap(scope.Attributes.ToSlice())
		logAttrs := map[string]string{}
		record.WalkAttributes(func(kv otellog.KeyValue) bool {
			logAttrs[kv.Key] = logValue(kv.Value)
			return true
		})

		timestamp := record.Timestamp()
		if timestamp.IsZero() {
			timestamp = record.ObservedTimestamp()
		}
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		serviceName := resourceAttrs["service.name"]
		if v := logAttrs["service.name"]; v != "" {
			serviceName = v
		}

		out = append(out, OTELLogRecord{
			RunnerID:               firstNonEmpty(logAttrs["runner.id"], resourceAttrs["runner.id"]),
			RunnerGroupID:          firstNonEmpty(logAttrs["runner_group.id"], resourceAttrs["runner_group.id"]),
			RunnerJobID:            firstNonEmpty(logAttrs["runner_job.id"], resourceAttrs["runner_job.id"]),
			RunnerJobExecutionID:   firstNonEmpty(logAttrs["runner_job_execution.id"], resourceAttrs["runner_job_execution.id"]),
			RunnerJobExecutionStep: firstNonEmpty(logAttrs["runner_job_execution_step.name"], resourceAttrs["runner_job_execution_step.name"]),
			ResourceAttributes:     resourceAttrs,
			ResourceSchemaURL:      record.Resource().SchemaURL(),
			ScopeSchemaURL:         scope.SchemaURL,
			ScopeName:              scope.Name,
			ScopeVersion:           scope.Version,
			ScopeAttributes:        scopeAttrs,
			Timestamp:              timestamp,
			ServiceName:            serviceName,
			SeverityNumber:         int(record.Severity()),
			SeverityText:           severityText(record.Severity()),
			Body:                   logValue(record.Body()),
			TraceID:                record.TraceID().String(),
			SpanID:                 record.SpanID().String(),
			TraceFlags:             int(record.TraceFlags()),
			LogAttributes:          logAttrs,
		})
	}

	return e.client.WriteControlPlaneLogs(ctx, e.logStreamID, out)
}

func (e *logExporter) Shutdown(context.Context) error {
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()
	return nil
}

func (e *logExporter) ForceFlush(context.Context) error { return nil }

type traceExporter struct {
	client   Client
	runnerID string

	mu     sync.Mutex
	closed bool
}

func (e *traceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	closed := e.closed
	e.mu.Unlock()
	if closed || len(spans) == 0 {
		return nil
	}

	out := make([]OTELTraceRecord, 0, len(spans))
	for _, span := range spans {
		resourceAttrs := attributeMap(span.Resource().Attributes())
		scope := span.InstrumentationScope()
		spanAttrs := attributeMap(span.Attributes())
		serviceName := resourceAttrs["service.name"]

		events := span.Events()
		eventTimes := make([]time.Time, 0, len(events))
		eventNames := make([]string, 0, len(events))
		eventAttrs := make([]map[string]string, 0, len(events))
		for _, event := range events {
			eventTimes = append(eventTimes, event.Time)
			eventNames = append(eventNames, event.Name)
			eventAttrs = append(eventAttrs, attributeMap(event.Attributes))
		}

		links := span.Links()
		linkTraceIDs := make([]string, 0, len(links))
		linkSpanIDs := make([]string, 0, len(links))
		linkStates := make([]string, 0, len(links))
		linkAttrs := make([]map[string]string, 0, len(links))
		for _, link := range links {
			linkTraceIDs = append(linkTraceIDs, link.SpanContext.TraceID().String())
			linkSpanIDs = append(linkSpanIDs, link.SpanContext.SpanID().String())
			linkStates = append(linkStates, link.SpanContext.TraceState().String())
			linkAttrs = append(linkAttrs, attributeMap(link.Attributes))
		}

		parentSpanID := ""
		if parent := span.Parent(); parent.IsValid() {
			parentSpanID = parent.SpanID().String()
		}

		out = append(out, OTELTraceRecord{
			RunnerID:               e.runnerID,
			RunnerGroupID:          resourceAttrs["runner_group.id"],
			RunnerJobID:            spanAttrs["runner_job.id"],
			RunnerJobExecutionID:   spanAttrs["runner_job_execution.id"],
			RunnerJobExecutionStep: spanAttrs["runner_job_execution_step.name"],
			Timestamp:              span.StartTime(),
			ResourceAttributes:     resourceAttrs,
			ResourceSchemaURL:      span.Resource().SchemaURL(),
			ScopeSchemaURL:         scope.SchemaURL,
			ScopeName:              scope.Name,
			ScopeVersion:           scope.Version,
			ScopeAttributes:        attributeMap(scope.Attributes.ToSlice()),
			TraceID:                span.SpanContext().TraceID().String(),
			SpanID:                 span.SpanContext().SpanID().String(),
			ParentSpanID:           parentSpanID,
			TraceState:             span.SpanContext().TraceState().String(),
			SpanName:               span.Name(),
			SpanKind:               span.SpanKind().String(),
			ServiceName:            serviceName,
			SpanAttributes:         spanAttrs,
			Duration:               span.EndTime().Sub(span.StartTime()).Nanoseconds(),
			StatusCode:             span.Status().Code.String(),
			StatusMessage:          span.Status().Description,
			EventsTimestamp:        eventTimes,
			EventsName:             eventNames,
			EventsAttributes:       eventAttrs,
			LinksTraceID:           linkTraceIDs,
			LinksSpanID:            linkSpanIDs,
			LinksState:             linkStates,
			LinksAttributes:        linkAttrs,
		})
	}

	return e.client.WriteControlPlaneTraces(ctx, e.runnerID, out)
}

func (e *traceExporter) Shutdown(context.Context) error {
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()
	return nil
}

func attributeMap(attrs []attribute.KeyValue) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		out[string(attr.Key)] = attributeValue(attr.Value)
	}
	return out
}

func attributeValue(v attribute.Value) string {
	if v.Type() == attribute.STRING {
		return v.AsString()
	}
	return v.Emit()
}

func logValue(v otellog.Value) string {
	if v.Kind() == otellog.KindString {
		return v.AsString()
	}
	return v.String()
}

func severityText(severity otellog.Severity) string {
	switch {
	case severity >= otellog.SeverityTrace1 && severity <= otellog.SeverityTrace4:
		return "Trace"
	case severity >= otellog.SeverityDebug1 && severity <= otellog.SeverityDebug4:
		return "Debug"
	case severity >= otellog.SeverityInfo1 && severity <= otellog.SeverityInfo4:
		return "Info"
	case severity >= otellog.SeverityWarn1 && severity <= otellog.SeverityWarn4:
		return "Warn"
	case severity >= otellog.SeverityError1 && severity <= otellog.SeverityError4:
		return "Error"
	case severity >= otellog.SeverityFatal1 && severity <= otellog.SeverityFatal4:
		return "Fatal"
	default:
		return "Undefined"
	}
}

func firstNonEmpty(vals ...string) string {
	for _, val := range vals {
		if val != "" {
			return val
		}
	}
	return ""
}
