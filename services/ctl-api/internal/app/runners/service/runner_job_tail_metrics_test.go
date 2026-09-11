package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const tailMetricsTestRunnerID = "runnertest00000000000000001"

func TestTailRunnerJobsMetrics(t *testing.T) {
	tests := []struct {
		name          string
		wait          string
		probe         func(int, *gorm.DB)
		wake          bool
		status        int
		body          string
		handlerErrors int
		probes        map[string]int64
		sessions      map[string]int64
		wakes         int64
		probeQueries  int
	}{
		{
			name: "initial hit", wait: "1ms", status: http.StatusOK,
			probe: jobOnProbe(1), body: tailMetricsTestRunnerID,
			probes: map[string]int64{"hit": 1}, sessions: map[string]int64{jobTailOutcomeHotHit: 1}, probeQueries: 1,
		},
		{
			name: "notification woken hit", wait: "1s", wake: true, status: http.StatusOK,
			probe: jobOnProbe(2), body: tailMetricsTestRunnerID,
			probes: map[string]int64{"empty": 1, "hit": 1}, sessions: map[string]int64{jobTailOutcomeIdleThenHit: 1}, wakes: 1, probeQueries: 2,
		},
		{
			name: "backstop hit", wait: "7s", status: http.StatusOK,
			probe: jobOnProbe(2), body: tailMetricsTestRunnerID,
			probes: map[string]int64{"empty": 1, "hit": 1}, sessions: map[string]int64{jobTailOutcomeIdleThenHit: 1}, probeQueries: 2,
		},
		{
			name: "healthy empty timeout", wait: "2ms", status: http.StatusOK, body: "[]",
			probe:  func(_ int, _ *gorm.DB) { time.Sleep(10 * time.Millisecond) },
			probes: map[string]int64{"empty": 1}, sessions: map[string]int64{jobTailOutcomeTimeoutEmpty: 1}, probeQueries: 1,
		},
		{
			name: "transient probe failure recovered", wait: "1s", status: http.StatusOK,
			probe: func(n int, tx *gorm.DB) {
				if n == 1 {
					tx.AddError(context.DeadlineExceeded)
					return
				}
				jobOnProbe(2)(n, tx)
			}, body: tailMetricsTestRunnerID,
			probes: map[string]int64{"transient_error": 1, "hit": 1}, sessions: map[string]int64{jobTailOutcomeIdleThenHit: 1}, probeQueries: 2,
		},
		{
			name: "retry exhaustion", wait: "2ms", status: http.StatusServiceUnavailable,
			probe: func(_ int, tx *gorm.DB) { tx.AddError(context.DeadlineExceeded) }, body: "service unavailable",
			probes: map[string]int64{"transient_error": 1}, sessions: map[string]int64{jobTailOutcomeError: 1}, probeQueries: 1,
		},
		{
			name: "fatal probe failure", wait: "1s", status: http.StatusOK, handlerErrors: 1,
			probe:  func(_ int, tx *gorm.DB) { tx.AddError(errors.New("fatal query failure")) },
			probes: map[string]int64{"error": 1}, sessions: map[string]int64{jobTailOutcomeError: 1}, probeQueries: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTailHandlerHarness(t, tc.probe)
			defer h.shutdown()
			result := make(chan *httptest.ResponseRecorder, 1)
			go func() { result <- h.request(context.Background(), tc.wait) }()
			if tc.wake {
				require.Eventually(t, func() bool { return h.probeCount.Load() == 1 }, time.Second, time.Millisecond)
				require.Equal(t, 1, h.wake.Wake(tailMetricsTestRunnerID))
			}
			recorder := <-result
			require.Equal(t, tc.status, recorder.Code)
			require.Contains(t, recorder.Body.String(), tc.body)
			require.Equal(t, tc.handlerErrors, h.lastErrorCount)
			require.Equal(t, int64(tc.probeQueries), h.probeCount.Load())
			assertTailMetrics(t, h.reader, tc.sessions, tc.probes, tc.wakes)
		})
	}
}

func TestTailRunnerJobsCancellation(t *testing.T) {
	t.Run("during wait", func(t *testing.T) {
		h := newTailHandlerHarness(t, nil)
		defer h.shutdown()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan *httptest.ResponseRecorder, 1)
		go func() { result <- h.request(ctx, "1s") }()
		require.Eventually(t, func() bool { return h.probeCount.Load() == 1 }, time.Second, time.Millisecond)
		cancel()
		require.Equal(t, http.StatusOK, (<-result).Code)
		assertTailMetrics(t, h.reader, map[string]int64{jobTailOutcomeClientCancel: 1}, map[string]int64{"empty": 1}, 0)
	})

	t.Run("during probe semaphore", func(t *testing.T) {
		for range jobTailMaxConcurrentProbes {
			jobTailProbeSem <- struct{}{}
		}
		defer func() {
			for range jobTailMaxConcurrentProbes {
				<-jobTailProbeSem
			}
		}()
		h := newTailHandlerHarness(t, nil)
		defer h.shutdown()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan *httptest.ResponseRecorder, 1)
		go func() { result <- h.request(ctx, "1s") }()
		time.Sleep(5 * time.Millisecond)
		cancel()
		require.Equal(t, http.StatusOK, (<-result).Code)
		require.Zero(t, h.probeCount.Load())
		assertTailMetrics(t, h.reader, map[string]int64{jobTailOutcomeClientCancel: 1}, map[string]int64{"cancelled": 1}, 0)
	})
}

func TestTailRunnerJobsValidationEmitsNoSession(t *testing.T) {
	h := newTailHandlerHarness(t, nil)
	defer h.shutdown()
	recorder := h.request(context.Background(), "not-a-duration")
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1, h.lastErrorCount)
	require.Zero(t, h.probeCount.Load())
	assertTailMetrics(t, h.reader, nil, nil, 0)
}

func TestRunnerJobListenerConnectedGaugeAndAtomicObservation(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { require.NoError(t, provider.Shutdown(context.Background())) }()
	m := newRunnerJobListenerMetrics(provider)
	assertGauge := func(want int64) {
		var data metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &data))
		require.Equal(t, want, metricGauge(t, data, "nuon.runner.job_tail.listener.connected"))
	}
	assertGauge(0)
	m.setConnected(1)
	assertGauge(1)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(value int64) { defer wg.Done(); m.setConnected(value) }(int64(i % 2))
	}
	wg.Wait()
	require.Contains(t, []int64{0, 1}, m.connected.Load())

	require.Nil(t, newRunnerJobTailMetrics(nil))
	require.Nil(t, newRunnerJobListenerMetrics(nil))
	(*runnerJobTailMetrics)(nil).session(jobTailOutcomeError)
	(*runnerJobTailMetrics)(nil).probe(context.Background(), nil, nil)
	(*runnerJobTailMetrics)(nil).wake(context.Background())
	(*runnerJobListenerMetrics)(nil).setConnected(1)
	(*runnerJobListenerMetrics)(nil).failure(context.Background())
	(*runnerJobListenerMetrics)(nil).notification(context.Background(), "valid")
}

type tailHandlerHarness struct {
	service        *service
	reader         *sdkmetric.ManualReader
	provider       *sdkmetric.MeterProvider
	wake           *RunnerJobWakeRegistry
	probeCount     atomic.Int64
	lastErrorCount int
}

func newTailHandlerHarness(t *testing.T, probe func(int, *gorm.DB)) *tailHandlerHarness {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused"}), &gorm.Config{DisableAutomaticPing: true, DryRun: true})
	require.NoError(t, err)
	h := &tailHandlerHarness{reader: sdkmetric.NewManualReader(), wake: NewRunnerJobWakeRegistry()}
	h.provider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(h.reader))
	require.NoError(t, db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
		switch dest := tx.Statement.Dest.(type) {
		case *app.Runner:
			dest.ID = tailMetricsTestRunnerID
			dest.StatusV2.Metadata = map[string]any{}
			tx.RowsAffected = 1
		case *app.RunnerJob:
			n := int(h.probeCount.Add(1))
			if probe != nil {
				probe(n, tx)
			}
			if tx.Error == nil && tx.RowsAffected == 0 {
				tx.AddError(gorm.ErrRecordNotFound)
			}
		default:
			t.Fatalf("unexpected query destination %T", dest)
		}
	}))
	mw, err := metrics.New(validator.New(), metrics.WithDisable(true), metrics.WithLogger(zap.NewNop()))
	require.NoError(t, err)
	h.service = &service{db: db, l: zap.NewNop(), mw: mw, tailMetrics: newRunnerJobTailMetrics(h.provider), runnerJobWake: h.wake}
	return h
}

func (h *tailHandlerHarness) request(ctx context.Context, wait string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	ginCtx.Params = gin.Params{{Key: "runner_id", Value: tailMetricsTestRunnerID}}
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/v1/runners/"+tailMetricsTestRunnerID+"/jobs/tail?wait="+wait, nil).WithContext(ctx)
	h.service.TailRunnerJobs(ginCtx)
	h.lastErrorCount = len(ginCtx.Errors)
	return recorder
}

func (h *tailHandlerHarness) shutdown() { _ = h.provider.Shutdown(context.Background()) }

func jobOnProbe(want int) func(int, *gorm.DB) {
	return func(n int, tx *gorm.DB) {
		if n != want {
			return
		}
		job := tx.Statement.Dest.(*app.RunnerJob)
		job.ID = "runnerjobtest0000000000001"
		job.RunnerID = tailMetricsTestRunnerID
		job.Status = app.RunnerJobStatusAvailable
		tx.RowsAffected = 1
	}
}

func assertTailMetrics(t *testing.T, reader *sdkmetric.ManualReader, sessions, probes map[string]int64, wakes int64) {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	assertOutcomePopulation(t, data, "nuon.runner.job_tail.sessions", []string{jobTailOutcomeHotHit, jobTailOutcomeIdleThenHit, jobTailOutcomeTimeoutEmpty, jobTailOutcomeClientCancel, jobTailOutcomeError}, sessions)
	assertOutcomePopulation(t, data, "nuon.runner.job_tail.probes", []string{"hit", "empty", "transient_error", "error", "cancelled"}, probes)
	metric := metricByName(t, data, "nuon.runner.job_tail.notification.wakes")
	points := metric.Data.(metricdata.Sum[int64]).DataPoints
	require.Len(t, points, 1)
	require.Empty(t, points[0].Attributes.ToSlice())
	require.Equal(t, wakes, points[0].Value)
}

func assertOutcomePopulation(t *testing.T, data metricdata.ResourceMetrics, name string, outcomes []string, want map[string]int64) {
	t.Helper()
	points := metricByName(t, data, name).Data.(metricdata.Sum[int64]).DataPoints
	require.Len(t, points, len(outcomes))
	got := make(map[string]int64, len(points))
	for _, point := range points {
		attrs := point.Attributes.ToSlice()
		require.Len(t, attrs, 1)
		require.Equal(t, "outcome", string(attrs[0].Key))
		got[attrs[0].Value.AsString()] = point.Value
	}
	for _, outcome := range outcomes {
		require.Contains(t, got, outcome)
		require.Equal(t, want[outcome], got[outcome], outcome)
	}
}

func metricByName(t *testing.T, data metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, scope := range data.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name == name {
				return metric
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return metricdata.Metrics{}
}

func metricGauge(t *testing.T, data metricdata.ResourceMetrics, name string) int64 {
	t.Helper()
	points := metricByName(t, data, name).Data.(metricdata.Gauge[int64]).DataPoints
	require.Len(t, points, 1)
	return points[0].Value
}
