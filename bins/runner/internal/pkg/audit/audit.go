package audit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/otelresource"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/version"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	AttrAudit      = "nuon.audit"
	AttrAuditValue = "true"
	AttrEvent      = "nuon.audit.event"
	AttrOutcome    = "nuon.audit.outcome"

	OutcomeStarted   = "started"
	OutcomeSucceeded = "succeeded"
	OutcomeFailed    = "failed"
	OutcomeStopping  = "stopping"

	asyncExportTimeout = 10 * time.Second
	SyncExportTimeout  = 3 * time.Second
)

const (
	AsyncRouteAddress = "127.0.0.1:14318"
	SyncRouteAddress  = "127.0.0.1:14319"
)

var ErrUnavailable = errors.New("customer audit export is unavailable")

var jobEventTypes = map[models.AppRunnerJobGroup]string{
	models.AppRunnerJobGroupDeploy:  "install_deploy",
	models.AppRunnerJobGroupActions: "install_action_workflow_run",
	models.AppRunnerJobGroupSandbox: "install_sandbox_run",
}

type Event struct {
	Name       string
	Message    string
	Outcome    string
	Timestamp  time.Time
	Attributes map[string]string
}

type exporter interface {
	Export(context.Context, []sdklog.Record) error
	Shutdown(context.Context) error
}

type metricsWriter interface {
	Incr(string, []string)
	Timing(string, time.Duration, []string)
}

type syncExportRequest struct {
	exporter exporter
	err      error
}

type syncExportRequestKey struct{}

type syncProcessor struct{}

func (syncProcessor) Enabled(context.Context, sdklog.EnabledParameters) bool {
	return true
}

func (syncProcessor) OnEmit(ctx context.Context, record *sdklog.Record) error {
	request, _ := ctx.Value(syncExportRequestKey{}).(*syncExportRequest)
	if request == nil {
		return nil
	}
	request.err = request.exporter.Export(ctx, []sdklog.Record{record.Clone()})
	return request.err
}

func (syncProcessor) Shutdown(context.Context) error {
	return nil
}

func (syncProcessor) ForceFlush(context.Context) error {
	return nil
}

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Registrar *process.Registrar
	Settings  *settings.Settings
	Metrics   metrics.Writer
}

type Writer struct {
	otel     otellog.Logger
	identity map[string]string
	metrics  metricsWriter

	mu            sync.RWMutex
	enabled       bool
	syncTimeout   time.Duration
	asyncExporter exporter
	syncExporter  exporter

	starting struct {
		sync.Mutex
		complete bool
	}
	stopping struct {
		sync.Mutex
		complete bool
	}
}

func New(params Params) (*Writer, error) {
	asyncExporter, err := newOTLPExporter(context.Background(), "http://"+AsyncRouteAddress)
	if err != nil {
		return nil, fmt.Errorf("create asynchronous customer audit exporter: %w", err)
	}
	syncExporter, err := newOTLPExporter(context.Background(), "http://"+SyncRouteAddress)
	if err != nil {
		shutdownExporter(asyncExporter)
		return nil, fmt.Errorf("create synchronous customer audit exporter: %w", err)
	}
	w := newWriter(processIdentity(params.Registrar.ProcessID(), params.Registrar.ProcessType(), params.Settings), params.Metrics, asyncExporter, syncExporter, otelresource.New(params.Settings, ""))
	params.Lifecycle.Append(fx.Hook{
		OnStop: w.stop,
	})
	return w, nil
}

func newWriter(identity map[string]string, metrics metricsWriter, asyncExporter, syncExporter exporter, rsrc *sdkresource.Resource) *Writer {
	if rsrc == nil {
		rsrc = sdkresource.Empty()
	}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(rsrc),
		sdklog.WithProcessor(syncProcessor{}),
	)
	return &Writer{
		otel:          provider.Logger("github.com/nuonco/nuon/bins/runner/audit"),
		identity:      identity,
		metrics:       metrics,
		syncTimeout:   SyncExportTimeout,
		asyncExporter: asyncExporter,
		syncExporter:  syncExporter,
	}
}

func processIdentity(processID, processType string, settings *settings.Settings) map[string]string {
	identity := map[string]string{
		"runner_process.id":      processID,
		"runner_process.type":    processType,
		"runner_process.version": version.Version,
		"runner.id":              settings.Metadata["runner.id"],
		"runner.type":            settings.Metadata["runner.type"],
		"runner.platform":        settings.Platform,
		"install.id":             settings.Metadata["install.id"],
		"org.id":                 settings.Metadata["org.id"],
	}
	if platform := settings.Metadata["runner.platform"]; platform != "" {
		identity["runner.platform"] = platform
	}
	if settings.ContainerImageURL != "" {
		identity["runner.image.configured_url"] = settings.ContainerImageURL
	}
	if settings.ContainerImageTag != "" {
		identity["runner.image.configured_tag"] = settings.ContainerImageTag
		identity["runner.image.tag_identity"] = "configured"
	}
	return identity
}

func newOTLPExporter(ctx context.Context, baseEndpoint string) (exporter, error) {
	endpoint := strings.TrimRight(baseEndpoint, "/") + "/v1/logs"
	return otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithRetry(otlploghttp.RetryConfig{Enabled: false}),
	)
}

func (w *Writer) Enable() error {
	w.mu.Lock()
	w.enabled = true
	w.mu.Unlock()

	w.starting.Lock()
	defer w.starting.Unlock()
	if w.starting.complete {
		return nil
	}
	if err := w.WriteSync(context.Background(), Event{Name: "runner_process_lifecycle", Message: "runner process started", Outcome: OutcomeStarted}); err != nil {
		return err
	}
	w.starting.complete = true
	return nil
}

func (w *Writer) Disable() {
	w.mu.Lock()
	w.enabled = false
	w.mu.Unlock()
}

func (w *Writer) Available() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.enabled
}

func (w *Writer) WriteAsync(event Event) (err error) {
	started := time.Now()
	defer func() { w.observeWrite("async", started, err) }()
	w.mu.RLock()
	defer w.mu.RUnlock()
	if !w.enabled || w.asyncExporter == nil {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), asyncExportTimeout)
	defer cancel()
	if err := w.export(ctx, w.asyncExporter, event); err != nil {
		return fmt.Errorf("enqueue customer audit event through asynchronous route: %w", err)
	}
	return nil
}

func (w *Writer) WriteSync(ctx context.Context, event Event) (err error) {
	started := time.Now()
	defer func() { w.observeWrite("sync", started, err) }()
	w.mu.RLock()
	defer w.mu.RUnlock()
	if !w.enabled || w.syncExporter == nil {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, w.syncTimeout)
	defer cancel()
	if err := w.export(ctx, w.syncExporter, event); err != nil {
		return fmt.Errorf("export customer audit event: %w", err)
	}
	return nil
}

func (w *Writer) observeWrite(route string, started time.Time, err error) {
	if w.metrics == nil {
		return
	}
	status := "success"
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		status = "timeout"
	case errors.Is(err, context.Canceled):
		status = "canceled"
	case errors.Is(err, ErrUnavailable):
		status = "unavailable"
	case err != nil:
		status = "failure"
	}
	tags := []string{"route:" + route, "status:" + status}
	w.metrics.Incr("runner.audit_export.write", tags)
	w.metrics.Timing("runner.audit_export.write.duration", time.Since(started), tags)
}

func (w *Writer) export(ctx context.Context, destination exporter, event Event) error {
	request := &syncExportRequest{exporter: destination}
	w.otel.Emit(context.WithValue(ctx, syncExportRequestKey{}, request), w.record(event))
	return request.err
}

func (w *Writer) ProcessStopping(ctx context.Context, reason, source string) error {
	w.stopping.Lock()
	defer w.stopping.Unlock()
	if w.stopping.complete {
		return nil
	}
	err := w.WriteSync(ctx, Event{
		Name:    "runner_process_lifecycle",
		Message: "runner process stopping",
		Outcome: OutcomeStopping,
		Attributes: map[string]string{
			"runner.shutdown.reason": reason,
			"runner.shutdown.source": source,
		},
	})
	if err != nil {
		return err
	}
	w.stopping.complete = true
	return nil
}

func (w *Writer) record(event Event) otellog.Record {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	var record otellog.Record
	record.SetTimestamp(event.Timestamp)
	record.SetObservedTimestamp(time.Now().UTC())
	record.SetSeverity(otellog.SeverityInfo)
	record.SetSeverityText("INFO")
	record.SetEventName(event.Name)
	record.SetBody(otellog.StringValue(event.Message))
	record.AddAttributes(
		otellog.String(AttrAudit, AttrAuditValue),
		otellog.String(AttrEvent, event.Name),
		otellog.String(AttrOutcome, event.Outcome),
	)
	attributes := make(map[string]string, len(w.identity)+len(event.Attributes))
	for key, value := range w.identity {
		attributes[key] = value
	}
	for key, value := range event.Attributes {
		attributes[key] = value
	}
	for key, value := range attributes {
		if value != "" {
			record.AddAttributes(otellog.String(key, value))
		}
	}
	return record
}

func shutdownExporter(exporter exporter) {
	ctx, cancel := context.WithTimeout(context.Background(), SyncExportTimeout)
	defer cancel()
	_ = exporter.Shutdown(ctx)
}

func (w *Writer) stop(ctx context.Context) error {
	w.mu.Lock()
	w.enabled = false
	asyncExporter := w.asyncExporter
	syncExporter := w.syncExporter
	w.asyncExporter = nil
	w.syncExporter = nil
	w.mu.Unlock()
	if asyncExporter != nil {
		shutdownExporter(asyncExporter)
	}
	if syncExporter != nil {
		shutdownExporter(syncExporter)
	}
	return nil
}

func JobEvent(job *models.AppRunnerJob, message, outcome string, attributes map[string]string) (Event, bool) {
	if job == nil || jobEventTypes[job.Group] == "" {
		return Event{}, false
	}
	attrs := make(map[string]string)
	for _, attr := range jobs.AuditAttrs(job) {
		attrs[string(attr.Key)] = attr.Value.AsString()
	}
	for key, value := range attributes {
		attrs[key] = value
	}
	if job.LogStreamID != "" {
		attrs["log_stream.id"] = job.LogStreamID
	}
	return Event{Name: jobEventTypes[job.Group], Message: message, Outcome: outcome, Attributes: attrs}, true
}
