package log

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/log-streams/%s/logs"
)

func NewOTELProvider(logStream *app.LogStream) (*log.LoggerProvider, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	url := fmt.Sprintf(defaultOTLPLogsEndpointTmpl, logStream.RunnerAPIURL, logStream.ID)

	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(url),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + logStream.WriteToken,
		}),
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("unable to initialize otlp log exporter: %w", err)
	}

	rsrc := getResource(logStream.ID, generics.ToStringMap(logStream.Attrs))
	lp := log.NewLoggerProvider(
		log.WithResource(rsrc),
		log.WithProcessor(
			log.NewBatchProcessor(logExporter),
		),
	)

	return lp, nil
}
