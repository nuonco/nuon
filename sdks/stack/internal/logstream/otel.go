// Package logstream wires a *slog.Logger to either stdout or the Nuon
// ctl-api OTLP log-stream ingest endpoint.
//
// The OTLP path mirrors bins/runner/internal/pkg/slog/otel.go: batched
// otlploghttp exporter posting to {RunnerAPIURL}/v1/log-streams/{id}/logs
// with a Bearer write token.
package logstream

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

const otlpEndpointTmpl = "%s/v1/log-streams/%s/logs"

// Config configures an OTLP-backed slog logger.
type Config struct {
	RunnerAPIURL string
	LogStreamID  string
	WriteToken   string
	// ServiceName is set on the OTEL resource (e.g. "stack").
	ServiceName string
	// Attrs are added as resource attributes (e.g. install_id).
	Attrs map[string]string
}

// Provider holds an OTEL log provider that must be shut down when done.
//
// Logger() emits records under the OTEL instrumentation scope "oteljob" — the
// runner's convention for user-visible job output, which the dashboard surfaces
// by default. SystemLogger() currently aliases Logger() (everything emits as
// oteljob); the accessor is preserved so call sites can be migrated to a
// real "system" scope iteratively as we validate which records belong in the
// hidden-by-default bucket.
type Provider struct {
	lp     *sdklog.LoggerProvider
	user   *slog.Logger
	system *slog.Logger
}

func (p *Provider) Logger() *slog.Logger       { return p.user }
func (p *Provider) SystemLogger() *slog.Logger { return p.system }

func (p *Provider) Shutdown(ctx context.Context) error {
	if p.lp == nil {
		return nil
	}
	return p.lp.Shutdown(ctx)
}

// teeHandler fans out each record to two slog handlers. Used so the CLI shows
// progress on the customer's terminal while we also push to the dashboard.
type teeHandler struct{ a, b slog.Handler }

func (h teeHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.a.Enabled(ctx, l) || h.b.Enabled(ctx, l)
}

func (h teeHandler) Handle(ctx context.Context, r slog.Record) error {
	// Best-effort — a failure on one sink shouldn't suppress the other.
	errA := h.a.Handle(ctx, r.Clone())
	errB := h.b.Handle(ctx, r)
	if errA != nil {
		return errA
	}
	return errB
}

func (h teeHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return teeHandler{a: h.a.WithAttrs(as), b: h.b.WithAttrs(as)}
}

func (h teeHandler) WithGroup(name string) slog.Handler {
	return teeHandler{a: h.a.WithGroup(name), b: h.b.WithGroup(name)}
}

// NewStdout returns a Provider that writes to stdout. Useful for local dev.
// Both Logger() and SystemLogger() return the same stdout logger — the
// oteljob/system split only matters once records reach the dashboard.
func NewStdout(serviceName string) *Provider {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h).With("service", serviceName)
	return &Provider{user: l, system: l}
}

// New creates an OTLP-backed Provider. Caller must call Shutdown.
func New(ctx context.Context, cfg Config) (*Provider, error) {
	if cfg.RunnerAPIURL == "" || cfg.LogStreamID == "" || cfg.WriteToken == "" {
		return nil, fmt.Errorf("logstream: RunnerAPIURL, LogStreamID, WriteToken all required")
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "stack"
	}

	endpoint := fmt.Sprintf(otlpEndpointTmpl, cfg.RunnerAPIURL, cfg.LogStreamID)
	exp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.WriteToken,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("logstream: init otlp exporter: %w", err)
	}

	attrs := []attribute.KeyValue{attribute.String("service.name", cfg.ServiceName)}
	for k, v := range cfg.Attrs {
		attrs = append(attrs, attribute.String(k, v))
	}
	rsrc := resource.NewSchemaless(attrs...)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(rsrc),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exp,
			sdklog.WithExportMaxBatchSize(25),
			sdklog.WithExportInterval(5*time.Second),
		)),
	)

	// One OTEL logger under scope "oteljob" — the runner's convention for
	// user-visible job output, which the dashboard surfaces by default. Each
	// CLI invocation is shaped like a single runner job (short-lived, no
	// long-running process), so for now everything emits under oteljob.
	// SystemLogger aliases this logger; individual call sites will migrate to
	// a real "system"-scoped logger iteratively as we validate which records
	// belong in the hidden-by-default bucket.
	otelLogger := otelslog.NewLogger("oteljob", otelslog.WithLoggerProvider(lp))

	// Tee through stdout so the customer running the CLI also sees progress.
	stdoutH := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(teeHandler{a: stdoutH, b: otelLogger.Handler()}).With("service", cfg.ServiceName)
	return &Provider{lp: lp, user: logger, system: logger}, nil
}
