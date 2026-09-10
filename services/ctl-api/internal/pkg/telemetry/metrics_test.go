package telemetry

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx/fxtest"
)

func TestMeterProviderWithoutEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "not-a-url")
	lc := fxtest.NewLifecycle(t)
	cfg, err := NewConfig(&internal.Config{})
	require.NoError(t, err)
	provider, err := NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	require.IsType(t, noop.NewMeterProvider(), provider)
	lc.RequireStart().RequireStop()
}

func TestMeterProviderExportsAndFlushesOnShutdown(t *testing.T) {
	type request struct {
		path, authorization string
		body                []byte
	}
	requests := make(chan request, 8)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- request{r.URL.Path, r.Header.Get("Authorization"), body}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", receiver.URL+"/tenant/")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_HEADERS", "")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer%20test-token")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	t.Setenv("OTEL_METRICS_EXEMPLAR_FILTER", "always_on")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "base2_exponential_bucket_histogram")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "nuon.control_plane.id=cp-test,service.instance.id=replica-a")
	t.Setenv("OTEL_SERVICE_NAME", "control-plane-api")
	lc := fxtest.NewLifecycle(t)
	globalMeter, globalTracer := otel.GetMeterProvider(), otel.GetTracerProvider()
	serviceConfig := &internal.Config{
		DisableMetrics:    true,
		ServiceName:       "ctl-api",
		ServiceType:       "api",
		ServiceDeployment: "public",
		Version:           "test-version",
	}
	require.NoError(t, config.NewFileLoader(filepath.Join(t.TempDir(), "config.yaml")).LoadInto(nil, serviceConfig))
	cfg, err := NewConfig(serviceConfig)
	require.NoError(t, err)
	provider, err := NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	require.Same(t, globalMeter, otel.GetMeterProvider())
	require.Same(t, globalTracer, otel.GetTracerProvider())
	lc.RequireStart()
	stopped := false
	t.Cleanup(func() {
		if !stopped {
			lc.RequireStop()
		}
	})
	m, err := metrics.NewHTTPMetrics(provider)
	require.NoError(t, err)
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled,
	}))
	m.Start(ctx, "public", "GET", "https")("/v1/apps/:id", 200)
	require.NoError(t, provider.(*sdkmetric.MeterProvider).ForceFlush(ctx))
	check := func(wantCount uint64) {
		t.Helper()
		select {
		case req := <-requests:
			require.Equal(t, "/tenant/v1/metrics", req.path)
			require.Equal(t, "Bearer test-token", req.authorization)
			payload := pmetricotlp.NewExportRequest()
			require.NoError(t, payload.UnmarshalProto(req.body))
			rms := payload.Metrics().ResourceMetrics()
			require.Equal(t, 1, rms.Len())
			attrs := rms.At(0).Resource().Attributes().AsRaw()
			require.Equal(t, "cp-test", attrs["nuon.control_plane.id"])
			require.Equal(t, "replica-a", attrs["service.instance.id"])
			require.Equal(t, "control-plane-api", attrs["service.name"])
			require.Equal(t, "test-version", attrs["service.version"])
			metrics := rms.At(0).ScopeMetrics().At(0).Metrics()
			found := false
			for i := 0; i < metrics.Len(); i++ {
				metric := metrics.At(i)
				if metric.Name() == "http.server.request.duration" {
					found = true
					require.Equal(t, "s", metric.Unit())
					require.Equal(t, pmetric.AggregationTemporalityCumulative, metric.Histogram().AggregationTemporality())
					points := metric.Histogram().DataPoints()
					require.Equal(t, 1, points.Len())
					require.Equal(t, wantCount, points.At(0).Count())
					require.Zero(t, points.At(0).Exemplars().Len())
				}
			}
			require.True(t, found)
		case <-time.After(time.Second):
			t.Fatal("no OTLP metric export received")
		}
	}
	check(1)
	m.Start(ctx, "public", "GET", "https")("/v1/apps/:id", 200)
	lc.RequireStop()
	stopped = true
	check(2)
}

func TestSlowMetricExporterDoesNotBlockHTTPRequests(t *testing.T) {
	started := make(chan struct{})
	var first sync.Once
	var unavailable atomic.Bool
	unavailable.Store(true)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if unavailable.Load() {
			first.Do(func() { close(started) })
			<-r.Context().Done()
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	lc := fxtest.NewLifecycle(t)
	cfg, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL, ServiceName: "ctl-api"})
	require.NoError(t, err)
	provider, err := NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	lc.RequireStart()
	t.Cleanup(func() { lc.RequireStop() })
	m, err := metrics.NewHTTPMetrics(provider)
	require.NoError(t, err)
	m.Start(context.Background(), "runner", "GET", "http")("/v1/jobs", 200)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	flushed := make(chan error, 1)
	go func() { flushed <- provider.(*sdkmetric.MeterProvider).ForceFlush(ctx) }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("export did not reach receiver")
	}
	completed := make(chan int, 1)
	go func() {
		handler := m.Handler("runner", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.SetHTTPRoute(r.Context(), "/v1/jobs")
			w.WriteHeader(http.StatusAccepted)
		}))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/v1/jobs", nil))
		completed <- response.Code
	}()
	select {
	case code := <-completed:
		require.Equal(t, http.StatusAccepted, code)
	case <-time.After(time.Second):
		t.Fatal("request blocked on telemetry export")
	}
	cancel()
	require.Error(t, <-flushed)
	unavailable.Store(false)
	require.NoError(t, provider.(*sdkmetric.MeterProvider).ForceFlush(context.Background()))
}

func TestGenericTLSHeadersAndCompression(t *testing.T) {
	requests := make(chan []byte, 2)
	receiver := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer a+b,c=token" || r.Header.Get("Content-Encoding") != "gzip" || len(r.TLS.PeerCertificates) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer reader.Close()
		body, _ := io.ReadAll(reader)
		requests <- body
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	receiver.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert}
	receiver.StartTLS()
	t.Cleanup(receiver.Close)
	cert := receiver.TLS.Certificates[0]
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	require.NoError(t, err)
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0600))
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600))
	t.Setenv("OTEL_EXPORTER_OTLP_CERTIFICATE", certPath)
	t.Setenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE", certPath)
	t.Setenv("OTEL_EXPORTER_OTLP_CLIENT_KEY", keyPath)
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer%20a+b%2Cc=token")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "gzip")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "false")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	lc := fxtest.NewLifecycle(t)
	cfg, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL})
	require.NoError(t, err)
	provider, err := NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	lc.RequireStart()
	t.Cleanup(func() { lc.RequireStop() })
	counter, err := provider.Meter("test").Int64Counter("test.count")
	require.NoError(t, err)
	counter.Add(context.Background(), 7)
	require.NoError(t, provider.(*sdkmetric.MeterProvider).ForceFlush(context.Background()))
	payload := pmetricotlp.NewExportRequest()
	require.NoError(t, payload.UnmarshalProto(<-requests))
	require.EqualValues(t, 7, payload.Metrics().ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).IntValue())
}

func TestExportErrorDetailsAndRecovery(t *testing.T) {
	var reject atomic.Bool
	reject.Store(true)
	requests := make(chan []byte, 2)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if reject.Load() {
			http.Error(w, "collector authentication rejected", http.StatusUnauthorized)
			return
		}
		requests <- body
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	lc := fxtest.NewLifecycle(t)
	cfg, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL + "/tenant"})
	require.NoError(t, err)
	provider, err := NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	lc.RequireStart()
	t.Cleanup(func() { reject.Store(false); lc.RequireStop() })
	counter, err := provider.Meter("test").Int64Counter("test.count")
	require.NoError(t, err)
	counter.Add(context.Background(), 3)
	err = provider.(*sdkmetric.MeterProvider).ForceFlush(context.Background())
	require.ErrorContains(t, err, "collector authentication rejected")
	require.ErrorContains(t, err, "401 Unauthorized")
	require.ErrorContains(t, err, receiver.URL+"/tenant/v1/metrics")
	reject.Store(false)
	counter.Add(context.Background(), 5)
	require.NoError(t, provider.(*sdkmetric.MeterProvider).ForceFlush(context.Background()))
	payload := pmetricotlp.NewExportRequest()
	require.NoError(t, payload.UnmarshalProto(<-requests))
	require.EqualValues(t, 8, payload.Metrics().ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).IntValue())
}

func TestExporterTimeoutAndShutdownDeadline(t *testing.T) {
	for _, tc := range []struct {
		name          string
		exportTimeout string
		stopTimeout   time.Duration
	}{
		{"export timeout", "50", time.Second},
		{"application deadline", "60000", 100 * time.Millisecond},
		{"provider deadline", "60000", 10 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var reached atomic.Int64
			receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				reached.Add(1)
				<-r.Context().Done()
			}))
			t.Cleanup(receiver.Close)
			t.Setenv("OTEL_EXPORTER_OTLP_TIMEOUT", tc.exportTimeout)
			t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
			lc := fxtest.NewLifecycle(t)
			cfg, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL})
			require.NoError(t, err)
			provider, err := NewMeterProvider(lc, cfg)
			require.NoError(t, err)
			lc.RequireStart()
			counter, err := provider.Meter("test").Int64Counter("test.count")
			require.NoError(t, err)
			counter.Add(context.Background(), 1)
			ctx, cancel := context.WithTimeout(context.Background(), tc.stopTimeout)
			defer cancel()
			started := time.Now()
			err = lc.Stop(ctx)
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.Less(t, time.Since(started), min(tc.stopTimeout, 5*time.Second)+time.Second)
			require.GreaterOrEqual(t, reached.Load(), int64(1))
		})
	}
}

func TestCardinalityOverflowPreservesCounts(t *testing.T) {
	requests := make(chan []byte, 4)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- body
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	lc := fxtest.NewLifecycle(t)
	cfg, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL})
	require.NoError(t, err)
	provider, err := NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	lc.RequireStart()
	t.Cleanup(func() { lc.RequireStop() })
	m, err := metrics.NewHTTPMetrics(provider)
	require.NoError(t, err)
	for batch := 0; batch < 2; batch++ {
		var wg sync.WaitGroup
		for worker := 0; worker < 4; worker++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				for i := 0; i < 1500; i++ {
					m.Start(context.Background(), "public", "GET", "http")(fmt.Sprintf("/route/%d/%d/%d", batch, worker, i), 200)
				}
			}(worker)
		}
		wg.Wait()
		require.NoError(t, provider.(*sdkmetric.MeterProvider).ForceFlush(context.Background()))
		payload := pmetricotlp.NewExportRequest()
		require.NoError(t, payload.UnmarshalProto(<-requests))
		observed := payload.Metrics().ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
		found := false
		for i := 0; i < observed.Len(); i++ {
			m := observed.At(i)
			if m.Name() == "http.server.active_requests" {
				require.Equal(t, 1, m.Sum().DataPoints().Len())
				require.Zero(t, m.Sum().DataPoints().At(0).IntValue())
			}
			if m.Name() != "http.server.request.duration" {
				continue
			}
			found = true
			points := m.Histogram().DataPoints()
			require.Equal(t, 2000, points.Len())
			var total, overflow uint64
			for j := 0; j < points.Len(); j++ {
				point := points.At(j)
				total += point.Count()
				if value, ok := point.Attributes().Get("otel.metric.overflow"); ok && value.Bool() {
					overflow += point.Count()
				}
			}
			require.EqualValues(t, (batch+1)*6000, total)
			require.EqualValues(t, (batch+1)*6000-1999, overflow)
		}
		require.True(t, found)
	}
}

func TestPeriodicExportSeparatesProcessResources(t *testing.T) {
	var mu sync.Mutex
	observed := map[string]int64{}
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		payload := pmetricotlp.NewExportRequest()
		if err := payload.UnmarshalProto(body); err != nil {
			w.WriteHeader(400)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		rms := payload.Metrics().ResourceMetrics()
		for i := 0; i < rms.Len(); i++ {
			rm := rms.At(i)
			cp, _ := rm.Resource().Attributes().Get("nuon.control_plane.id")
			if cp.Str() != "cp-test" {
				w.WriteHeader(400)
				return
			}
			id, _ := rm.Resource().Attributes().Get("service.instance.id")
			if rm.ScopeMetrics().Len() > 0 {
				observed[id.Str()] = rm.ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).IntValue()
			}
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "20")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "nuon.control_plane.id=cp-test")
	expected := map[string]int64{}
	for _, count := range []int64{7, 2} {
		lc := fxtest.NewLifecycle(t)
		cfg, err := NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL, ServiceName: "ctl-api"})
		require.NoError(t, err)
		id, _ := cfg.Resource.Set().Value("service.instance.id")
		expected[id.AsString()] = count
		provider, err := NewMeterProvider(lc, cfg)
		require.NoError(t, err)
		lc.RequireStart()
		t.Cleanup(func() { lc.RequireStop() })
		counter, err := provider.Meter("test").Int64Counter("test.count")
		require.NoError(t, err)
		counter.Add(context.Background(), count)
	}
	require.Len(t, expected, 2)
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return reflect.DeepEqual(expected, observed)
	}, 2*time.Second, 10*time.Millisecond)
}
