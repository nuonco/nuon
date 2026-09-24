package blobstore

import (
	"context"
	"io"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type operationMetrics struct {
	operations metric.Int64Counter
	duration   metric.Float64Histogram
}

func newOperationMetrics(provider metric.MeterProvider) *operationMetrics {
	if provider == nil {
		return nil
	}
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/blobstore")
	operations, err := meter.Int64Counter("nuon.blobstore.operations", metric.WithUnit("{operation}"))
	if err != nil {
		return nil
	}
	duration, err := meter.Float64Histogram("nuon.blobstore.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300))
	if err != nil {
		return nil
	}
	return &operationMetrics{operations: operations, duration: duration}
}

func (m *operationMetrics) record(ctx context.Context, operation string, started time.Time, outcome string) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("outcome", outcome),
	)
	m.operations.Add(ctx, 1, attrs)
	m.duration.Record(ctx, time.Since(started).Seconds(), attrs)
}

type instrumentedService struct {
	next    Service
	metrics *operationMetrics
}

func instrumentService(next Service, provider metric.MeterProvider) Service {
	metrics := newOperationMetrics(provider)
	if metrics == nil {
		return next
	}
	return &instrumentedService{next: next, metrics: metrics}
}

func (s *instrumentedService) Upload(ctx context.Context, key string, data []byte) (err error) {
	started := time.Now()
	defer func() { s.metrics.record(ctx, "write", started, outcome(err)) }()
	return s.next.Upload(ctx, key, data)
}

func (s *instrumentedService) Download(ctx context.Context, key string) (data []byte, err error) {
	started := time.Now()
	defer func() { s.metrics.record(ctx, "read", started, outcome(err)) }()
	return s.next.Download(ctx, key)
}

func (s *instrumentedService) UploadStream(ctx context.Context, key string, reader io.Reader) (checksum string, err error) {
	started := time.Now()
	defer func() { s.metrics.record(ctx, "write_stream", started, outcome(err)) }()
	return s.next.UploadStream(ctx, key, reader)
}

func (s *instrumentedService) DownloadStream(ctx context.Context, key string) (body io.ReadCloser, err error) {
	started := time.Now()
	body, err = s.next.DownloadStream(ctx, key)
	s.metrics.record(ctx, "read_stream_open", started, outcome(err))
	if err != nil {
		return nil, err
	}
	return &instrumentedReadCloser{ReadCloser: body, ctx: ctx, started: time.Now(), metrics: s.metrics}, nil
}

func (s *instrumentedService) GetMetadata(ctx context.Context, key string) (size int64, contentType string, err error) {
	started := time.Now()
	defer func() { s.metrics.record(ctx, "metadata", started, outcome(err)) }()
	return s.next.GetMetadata(ctx, key)
}

func outcome(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

type instrumentedReadCloser struct {
	io.ReadCloser
	ctx     context.Context
	started time.Time
	metrics *operationMetrics
	once    sync.Once
}

func (r *instrumentedReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err == io.EOF {
		r.record("success")
	} else if err != nil {
		r.record("error")
	}
	return n, err
}

func (r *instrumentedReadCloser) Close() error {
	err := r.ReadCloser.Close()
	if err != nil {
		r.record("error")
	} else {
		r.record("closed_early")
	}
	return err
}

func (r *instrumentedReadCloser) record(outcome string) {
	r.once.Do(func() { r.metrics.record(r.ctx, "read_stream_body", r.started, outcome) })
}
