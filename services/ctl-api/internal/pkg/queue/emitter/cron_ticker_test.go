package emitter

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"

	basemetrics "github.com/nuonco/nuon/pkg/metrics"
	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func TestCronTickerEmitsWhenSerializedEmitterOmitsSignalTemplate(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))

	emitter := &app.QueueEmitter{
		ID:         "emitter-id",
		QueueID:    "queue-id",
		Enabled:    true,
		SignalType: example.ExampleSignalType,
		SignalTemplate: signaldb.SignalData{
			Signal: &example.ExampleSignal{},
		},
	}
	env.OnActivity((*activities.Activities).GetEmitter, mock.Anything, mock.Anything, mock.Anything).
		Return(emitter, nil).Once()
	env.OnActivity((*activities.Activities).EmitSignal, mock.Anything, mock.Anything, mock.Anything).
		Return(&activities.EmitSignalResponse{QueueSignalID: "signal-id", WorkflowID: "workflow-id"}, nil).Once()
	env.OnActivity((*activities.Activities).UpdateSignalEmitter, mock.Anything, mock.Anything, mock.Anything).
		Return(&activities.UpdateSignalEmitterResponse{Success: true}, nil).Once()
	env.OnActivity((*activities.Activities).UpdateEmitterStats, mock.Anything, mock.Anything, mock.Anything).
		Return(&activities.UpdateEmitterStatsResponse{}, nil).Once()

	v := validator.New()
	mw, err := basemetrics.New(v, basemetrics.WithDisable(true))
	require.NoError(t, err)
	tmw, err := tmetrics.New(v, tmetrics.WithMetricsWriter(mw))
	require.NoError(t, err)

	env.ExecuteWorkflow((&Workflows{cfg: &internal.Config{}, mw: tmw}).CronTicker, CronTickerWorkflowRequest{
		QueueID:   emitter.QueueID,
		EmitterID: emitter.ID,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

// The parent only notices a disable on its next alive check, up to ten minutes
// later. Without this the ticker keeps emitting for an install whose runner is
// gone, which is the whole thing the disable exists to stop.
func TestCronTickerTerminatesWhenEmitterIsDisabled(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))

	emitter := &app.QueueEmitter{
		ID:             "emitter-id",
		QueueID:        "queue-id",
		Enabled:        false,
		DisabledReason: "no healthy runner",
		SignalType:     example.ExampleSignalType,
		SignalTemplate: signaldb.SignalData{
			Signal: &example.ExampleSignal{},
		},
	}
	env.OnActivity((*activities.Activities).GetEmitter, mock.Anything, mock.Anything, mock.Anything).
		Return(emitter, nil).Once()
	env.OnActivity((*activities.Activities).TerminateWorkflow, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	v := validator.New()
	mw, err := basemetrics.New(v, basemetrics.WithDisable(true))
	require.NoError(t, err)
	tmw, err := tmetrics.New(v, tmetrics.WithMetricsWriter(mw))
	require.NoError(t, err)

	env.ExecuteWorkflow((&Workflows{cfg: &internal.Config{}, mw: tmw}).CronTicker, CronTickerWorkflowRequest{
		QueueID:   emitter.QueueID,
		EmitterID: emitter.ID,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
