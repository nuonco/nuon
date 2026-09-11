package workflowmetrics

import (
	"context"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestExecutionCompletionBoundary(t *testing.T) {
	for _, tc := range []struct {
		name          string
		before, after app.Status
		want          bool
	}{
		{"success", app.StatusInProgress, app.StatusSuccess, true},
		{"failure", app.StatusInProgress, app.StatusError, true},
		{"cancel", app.StatusInProgress, app.StatusCancelled, true},
		{"validation failure", app.StatusPending, app.StatusError, true},
		{"retry park", app.StatusInProgress, app.StatusFailedPendingRetry, false},
		{"approval park", app.StatusInProgress, app.AwaitingApproval, false},
		{"repeated error", app.StatusError, app.StatusError, false},
		{"terminal correction", app.StatusError, app.StatusCancelled, false},
		{"late success", app.StatusCancelled, app.StatusSuccess, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			qs := app.QueueSignal{Type: "execute-workflow", OwnerID: "workflow", OwnerType: "install_workflows", Status: app.CompositeStatus{Status: tc.before}}
			require.Equal(t, tc.want, executionCompleted(qs, tc.after))
		})
	}
	for _, qs := range []app.QueueSignal{
		{Type: "other", OwnerID: "workflow", OwnerType: "install_workflows"},
		{Type: "execute-workflow", OwnerID: "workflow", OwnerType: "apps"},
		{Type: "execute-workflow", OwnerType: "install_workflows"},
		{Type: "execute-workflow", OwnerID: "workflow", OwnerType: "install_workflows", DeletedAt: 1},
	} {
		require.False(t, executionCompleted(qs, app.StatusSuccess))
	}
	for status, want := range map[app.Status]string{
		app.StatusSuccess: "success", app.StatusError: "error", app.StatusCancelled: "cancelled",
		app.StatusFailedPendingRetry: "unknown", app.StatusInProgress: "unknown", "future-state": "unknown",
	} {
		require.Equal(t, want, executionOutcome(status))
	}
}

func TestRetryDecisions(t *testing.T) {
	auto := app.CompositeStatus{Status: app.StatusError, Metadata: map[string]any{"auto_retried": true}}
	manual := app.CompositeStatus{Status: app.StatusDiscarded, Metadata: map[string]any{"retry_type": "manual"}}
	for _, tc := range []struct {
		name            string
		before, request app.CompositeStatus
		want            string
	}{
		{"auto", app.CompositeStatus{}, auto, "auto"},
		{"repeat auto", auto, auto, ""},
		{"manual", app.CompositeStatus{Status: app.StatusError}, manual, "manual"},
		{"repeat manual", manual, manual, ""},
		{"discarded without marker", app.CompositeStatus{Status: app.StatusDiscarded}, manual, ""},
		{"retained manual marker", app.CompositeStatus{Status: app.StatusError, Metadata: manual.Metadata}, manual, ""},
		{"inherited auto marker", auto, app.CompositeStatus{Status: app.StatusError}, ""},
		{"clone metadata", app.CompositeStatus{}, app.CompositeStatus{Status: app.StatusError, Metadata: map[string]any{"retry_type": "auto"}}, ""},
		{"initial clone", app.CompositeStatus{}, app.CompositeStatus{Status: app.StatusPending, Metadata: auto.Metadata}, ""},
		{"manual request is not a decision yet", app.CompositeStatus{}, app.CompositeStatus{Status: app.StatusError, Metadata: manual.Metadata}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := retrySource(tc.request)
			if !newRetryDecision(tc.before, source) {
				source = ""
			}
			require.Equal(t, tc.want, source)
		})
	}
}

func TestCounterInitializationAndDisabled(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	_, err := NewCounters(CounterParams{Config: &telemetry.Config{Endpoint: "http://collector"}, Provider: provider})
	require.NoError(t, err)
	metrics := collect(t, reader)
	require.NotContains(t, metrics, "nuon.workflow.start_delay", "no artificial zero-duration samples")
	require.NotContains(t, metrics, "nuon.workflow.elapsed_time")
	for name, points := range map[string]int{"nuon.workflow.executions.completed": 12, "nuon.workflow.step.retries": 6} {
		sum := metrics[name].Data.(metricdata.Sum[int64])
		require.True(t, sum.IsMonotonic)
		require.Equal(t, metricdata.CumulativeTemporality, sum.Temporality)
		require.Len(t, sum.DataPoints, points)
		for _, point := range sum.DataPoints {
			require.Zero(t, point.Value)
			require.Equal(t, 2, point.Attributes.Len())
		}
	}
	c, err := NewCounters(CounterParams{Config: &telemetry.Config{}})
	require.NoError(t, err)
	qs := app.QueueSignal{Type: "execute-workflow", OwnerID: "workflow", OwnerType: "install_workflows"}
	for _, counters := range []*Counters{nil, c} {
		counters.SignalStatusUpdated(context.Background(), qs, app.StatusSuccess)
		counters.StepStatusUpdated(context.Background(), app.WorkflowStep{InstallWorkflowID: "workflow"}, app.CompositeStatus{Status: app.StatusError, Metadata: map[string]any{"auto_retried": true}})
		counters.FlowStarted(context.Background(), app.Workflow{ID: "workflow", StartedAt: time.Now()})
	}
}

func TestWorkflowDurationBounds(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	c, err := NewCounters(CounterParams{Config: &telemetry.Config{Endpoint: "http://collector"}, Provider: provider})
	require.NoError(t, err)
	c.startDelay.Record(context.Background(), 30)
	c.startDelay.Record(context.Background(), 31)
	c.elapsedTime.Record(context.Background(), 86400)
	c.elapsedTime.Record(context.Background(), 90000)
	metrics := collect(t, reader)
	start := metrics["nuon.workflow.start_delay"].Data.(metricdata.Histogram[float64])
	require.Equal(t, "s", metrics["nuon.workflow.start_delay"].Unit)
	require.Equal(t, metricdata.CumulativeTemporality, start.Temporality)
	require.Len(t, start.DataPoints, 1)
	require.Equal(t, []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600}, start.DataPoints[0].Bounds)
	require.Equal(t, []uint64{0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0}, start.DataPoints[0].BucketCounts)
	require.Equal(t, uint64(2), start.DataPoints[0].Count)
	require.Equal(t, 61.0, start.DataPoints[0].Sum)
	elapsed := metrics["nuon.workflow.elapsed_time"].Data.(metricdata.Histogram[float64]).DataPoints[0]
	require.Equal(t, "s", metrics["nuon.workflow.elapsed_time"].Unit)
	require.Equal(t, []float64{10, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 21600, 86400}, elapsed.Bounds)
	require.Equal(t, []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}, elapsed.BucketCounts)
	require.Equal(t, 176400.0, elapsed.Sum)
}

func TestElapsedSeconds(t *testing.T) {
	start := time.Unix(1000, 0)
	for _, tc := range []struct {
		start, end time.Time
		seconds    float64
		valid      bool
	}{
		{start, start.Add(37*time.Second + 500*time.Millisecond), 37.5, true},
		{start, start, 0, true},
		{start, start.Add(-time.Second), 0, false},
		{time.Time{}, start, 0, false},
		{start, time.Time{}, 0, false},
	} {
		seconds, valid := elapsedSeconds(tc.start, tc.end)
		require.Equal(t, tc.valid, valid)
		if valid {
			require.Equal(t, tc.seconds, seconds)
		}
	}
}
