package audit

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	LC  fx.Lifecycle
}

// New builds the audit emitter. Its logger is deliberately standalone rather
// than teed into the system logger: audit records must not reach stderr, so
// ctl-api's normal log output stays byte-for-byte unchanged and the Datadog
// monitors reading it are unaffected.
func New(params Params) (*Emitter, error) {
	if params.Cfg.AuditOTLPEndpoint == "" {
		return &Emitter{}, nil
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpointURL(params.Cfg.AuditOTLPEndpoint),
	}
	if params.Cfg.AuditOTLPToken != "" {
		opts = append(opts, otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + params.Cfg.AuditOTLPToken,
		}))
	}

	exporter, err := otlploghttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize audit log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("ctl-api"),
		)),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	params.LC.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return lp.Shutdown(ctx)
		},
	})

	return NewEmitter(zap.New(otelzap.NewCore("audit", otelzap.WithLoggerProvider(lp)))), nil
}
