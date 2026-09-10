package log

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/log-streams/%s/logs"
)

func NewOTELProvider(logStream *app.LogStream) (*log.LoggerProvider, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	endpoint := fmt.Sprintf(defaultOTLPLogsEndpointTmpl, logStream.RunnerAPIURL, logStream.ID)
	u, err := url.Parse(endpoint)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("invalid log stream endpoint: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		cancelFn()
		return nil, fmt.Errorf("log stream endpoint must be an absolute HTTP(S) URL")
	}

	rsrc := getResource(logStream.ID, generics.ToStringMap(logStream.Attrs))
	// Explicit timeouts and limits mirror the OTel Logs SDK/exporter v0.18.0
	// defaults so OTEL environment overrides cannot change product log delivery.
	client := &http.Client{
		Transport: &logStreamTransport{next: http.DefaultTransport, resource: rsrc},
		Timeout:   10 * time.Second,
	}

	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + logStream.WriteToken,
		}),
		otlploghttp.WithTLSClientConfig(nil),
		otlploghttp.WithTimeout(client.Timeout),
		otlploghttp.WithCompression(otlploghttp.NoCompression),
		otlploghttp.WithHTTPClient(client),
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("unable to initialize otlp log exporter: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithResource(rsrc),
		log.WithAttributeCountLimit(128),
		log.WithAttributeValueLengthLimit(-1),
		log.WithProcessor(
			log.NewBatchProcessor(logExporter,
				log.WithMaxQueueSize(2048),
				log.WithExportMaxBatchSize(512),
				log.WithExportInterval(time.Second),
				log.WithExportTimeout(30*time.Second),
			),
		),
	)

	return lp, nil
}
