package metrics

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	duration         metric.Float64Histogram
	active           metric.Int64UpDownCounter
	declaredBodySize metric.Int64Histogram
	methods          map[string]bool
}

func NewHTTPMetrics(provider metric.MeterProvider) (*HTTPMetrics, error) {
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/http")
	duration, err := meter.Float64Histogram("http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of HTTP server requests."),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10),
	)
	if err != nil {
		return nil, err
	}
	active, err := meter.Int64UpDownCounter("http.server.active_requests",
		metric.WithUnit("{request}"),
		metric.WithDescription("Number of active HTTP server requests."),
	)
	if err != nil {
		return nil, err
	}
	declaredBodySize, err := meter.Int64Histogram("nuon.http.server.request.declared_body.size",
		metric.WithUnit("By"),
		metric.WithDescription("Declared HTTP request body size; unknown lengths are omitted."),
		metric.WithExplicitBucketBoundaries(0, 128, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216),
	)
	if err != nil {
		return nil, err
	}
	methods := "CONNECT,DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT,TRACE"
	if configured, ok := os.LookupEnv("OTEL_INSTRUMENTATION_HTTP_KNOWN_METHODS"); ok {
		methods = configured
	}
	m := &HTTPMetrics{duration: duration, active: active, declaredBodySize: declaredBodySize, methods: make(map[string]bool)}
	for _, method := range strings.Split(methods, ",") {
		if method = strings.TrimSpace(method); method != "" {
			m.methods[method] = true
		}
	}
	return m, nil
}

func (m *HTTPMetrics) Start(ctx context.Context, api, method, scheme string) func(string, int) {
	if !m.methods[method] {
		method = "_OTHER"
	}
	attrs := []attribute.KeyValue{
		attribute.String("nuon.api", api),
		attribute.String("http.request.method", method),
		attribute.String("url.scheme", scheme),
	}
	activeAttrs := metric.WithAttributes(attrs...)
	m.active.Add(ctx, 1, activeAttrs)
	started := time.Now()
	return func(route string, status int) {
		m.active.Add(ctx, -1, activeAttrs)
		attrs = append(attrs, attribute.Int("http.response.status_code", status))
		if route != "" {
			attrs = append(attrs, attribute.String("http.route", route))
		}
		if status >= http.StatusInternalServerError {
			attrs = append(attrs, attribute.String("error.type", strconv.Itoa(status)))
		}
		options := metric.WithAttributes(attrs...)
		m.duration.Record(ctx, time.Since(started).Seconds(), options)
		if request, ok := ctx.Value(httpRequestMetricsKey{}).(*httpRequestMetrics); ok {
			if request.declaredBodySize >= 0 {
				m.declaredBodySize.Record(ctx, request.declaredBodySize, options)
			}
		}
	}
}

type httpRequestMetricsKey struct{}

type httpRequestMetrics struct {
	route            string
	declaredBodySize int64
}

func SetHTTPRoute(ctx context.Context, template string) {
	if request, ok := ctx.Value(httpRequestMetricsKey{}).(*httpRequestMetrics); ok {
		request.route = template
	}
}

func (m *HTTPMetrics) Handler(api string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		request := &httpRequestMetrics{declaredBodySize: r.ContentLength}
		r = r.WithContext(context.WithValue(r.Context(), httpRequestMetricsKey{}, request))
		finish := m.Start(r.Context(), api, r.Method, scheme)
		status := http.StatusInternalServerError
		defer func() { finish(request.route, status) }()
		status = httpsnoop.CaptureMetrics(next, w, r).Code
	})
}
