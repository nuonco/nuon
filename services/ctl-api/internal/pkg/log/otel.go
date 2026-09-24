package log

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func NewOTELProvider(logStream *app.LogStream) (*log.LoggerProvider, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	endpoint := fmt.Sprintf("%s/v1/log-streams/%s/logs", logStream.RunnerAPIURL, logStream.ID)
	u, err := url.Parse(logStream.RunnerAPIURL)
	if err != nil {
		zap.L().Warn("invalid log stream endpoint; passing unnormalized URL to exporter", zap.String("log_stream_id", logStream.ID))
	} else {
		if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			zap.L().Warn("log stream endpoint is not an absolute HTTP(S) URL; log delivery may fail", zap.String("log_stream_id", logStream.ID))
		}
		endpoint = u.JoinPath("v1", "log-streams", logStream.ID, "logs").String()
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
