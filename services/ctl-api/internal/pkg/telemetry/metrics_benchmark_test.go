package telemetry

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
)

func BenchmarkHTTPMetrics(b *testing.B) {
	for _, mode := range []string{"absent", "healthy", "blocked"} {
		b.Run(mode, func(b *testing.B) {
			var blocked atomic.Bool
			blocked.Store(mode == "blocked")
			started := make(chan struct{})
			release := make(chan struct{})
			var once sync.Once
			receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				once.Do(func() { close(started) })
				if blocked.Load() {
					select {
					case <-r.Context().Done():
						return
					case <-release:
					}
				}
				w.Header().Set("Content-Type", "application/x-protobuf")
			}))
			defer receiver.Close()
			b.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "10")
			lc := fxtest.NewLifecycle(b)
			cfg := &internal.Config{}
			if mode != "absent" {
				cfg.OTELExporterOTLPEndpoint = receiver.URL
			}
			shared, err := NewConfig(cfg)
			require.NoError(b, err)
			provider, err := NewMeterProvider(lc, shared)
			require.NoError(b, err)
			lc.RequireStart()
			defer func() { blocked.Store(false); close(release); lc.RequireStop() }()
			m, err := metrics.NewHTTPMetrics(provider)
			require.NoError(b, err)
			m.Start(context.Background(), "public", "GET", "http")("/v1/apps/:id", 200)
			if mode != "absent" {
				select {
				case <-started:
				case <-time.After(time.Second):
					b.Fatal("periodic exporter did not start")
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				m.Start(context.Background(), "public", "GET", "http")("/v1/apps/:id", 200)
			}
			b.StopTimer()
		})
	}
}
