package signal

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.temporal.io/sdk/testsuite"
)

type lifecycleTestHook struct {
	name      string
	supports  bool
	decision  BeforePhaseDecision
	beforeErr error
	afterErr  error
}

func (h *lifecycleTestHook) Name() string                   { return h.name }
func (h *lifecycleTestHook) Supports(SignalPhaseEvent) bool { return h.supports }
func (h *lifecycleTestHook) BeforePhase(context.Context, SignalPhaseEvent) (BeforePhaseDecision, error) {
	return h.decision, h.beforeErr
}
func (h *lifecycleTestHook) AfterPhase(context.Context, SignalPhaseEvent, SignalPhaseOutcome) error {
	return h.afterErr
}

func TestSignalLifecycleActivityRecordsReturnedHookResults(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	hooks := []SignalLifecycleHook{
		&lifecycleTestHook{name: "workflow_lifecycle_webhook", supports: true, decision: AllowPhaseDecision()},
		&lifecycleTestHook{name: "unknown-hook", supports: true, decision: BeforePhaseDecision{Allow: false}},
		&lifecycleTestHook{name: "workflow_lifecycle_slack", supports: false},
	}
	activities := NewSignalLifecycleActivities(SignalLifecycleActivitiesParams{Hooks: hooks, MeterProvider: provider})
	event := SignalPhaseEvent{SignalType: "test", Phase: SignalPhaseExecute}

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities.RunSignalLifecycleBeforePhase)
	for range 2 {
		encoded, err := env.ExecuteActivity(activities.RunSignalLifecycleBeforePhase, &RunSignalLifecycleBeforePhaseRequest{Event: event})
		require.NoError(t, err)
		var response RunSignalLifecycleBeforePhaseResponse
		require.NoError(t, encoded.Get(&response))
		require.False(t, response.Allow)
	}

	points := collectHookInvocationPoints(t, reader)
	require.Len(t, points, 2)
	require.EqualValues(t, 2, points[attributeKey(
		attribute.String("nuon.event.hook.name", "workflow_lifecycle_webhook"),
		attribute.String("nuon.event.hook.phase", "execute"),
		attribute.String("nuon.event.hook.invocation", "before"),
		attribute.String("nuon.event.hook.outcome", "success"),
	)])
	require.EqualValues(t, 2, points[attributeKey(
		attribute.String("nuon.event.hook.name", "other"),
		attribute.String("nuon.event.hook.phase", "execute"),
		attribute.String("nuon.event.hook.invocation", "before"),
		attribute.String("nuon.event.hook.outcome", "blocked"),
	)])
}

func TestSignalLifecycleActivityRecordsErrorsIncludingSwallowedAfterError(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	hook := &lifecycleTestHook{
		name:      "flow_lifecycle_telemetry",
		supports:  true,
		decision:  AllowPhaseDecision(),
		beforeErr: errors.New("before failed"),
		afterErr:  errors.New("after failed"),
	}
	activities := NewSignalLifecycleActivities(SignalLifecycleActivitiesParams{Hooks: []SignalLifecycleHook{hook}, MeterProvider: provider})
	event := SignalPhaseEvent{SignalType: "test", Phase: SignalPhase("unexpected")}

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities.RunSignalLifecycleBeforePhase)
	env.RegisterActivity(activities.RunSignalLifecycleAfterPhase)
	_, err := env.ExecuteActivity(activities.RunSignalLifecycleBeforePhase, &RunSignalLifecycleBeforePhaseRequest{Event: event})
	require.Error(t, err)
	_, err = env.ExecuteActivity(activities.RunSignalLifecycleAfterPhase, &RunSignalLifecycleAfterPhaseRequest{Event: event, Outcome: SignalPhaseOutcome{Status: SignalStatusError}})
	require.NoError(t, err)

	points := collectHookInvocationPoints(t, reader)
	require.Len(t, points, 2)
	for _, invocation := range []string{"before", "after"} {
		require.EqualValues(t, 1, points[attributeKey(
			attribute.String("nuon.event.hook.name", "flow_lifecycle_telemetry"),
			attribute.String("nuon.event.hook.phase", "other"),
			attribute.String("nuon.event.hook.invocation", invocation),
			attribute.String("nuon.event.hook.outcome", "error"),
		)])
	}
}

func attributeKey(values ...attribute.KeyValue) attribute.Distinct {
	set := attribute.NewSet(values...)
	return set.Equivalent()
}

func collectHookInvocationPoints(t *testing.T, reader *sdkmetric.ManualReader) map[attribute.Distinct]int64 {
	t.Helper()
	var resourceMetrics metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &resourceMetrics))
	points := make(map[attribute.Distinct]int64)
	for _, scope := range resourceMetrics.ScopeMetrics {
		for _, measurement := range scope.Metrics {
			if measurement.Name != "nuon.event.hook.invocations" {
				continue
			}
			for _, point := range measurement.Data.(metricdata.Sum[int64]).DataPoints {
				points[point.Attributes.Equivalent()] = point.Value
			}
		}
	}
	return points
}
