package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handlerworkflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func TestQueueBacklogOverflow(t *testing.T) {
	for _, source := range []string{"recovery", "enqueue", "legacy-recovery"} {
		t.Run(source, func(t *testing.T) {
			const count = 70
			const capacity = 10
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetDataConverter(converter.NewCompositeDataConverter(
				converter.NewNilPayloadConverter(),
				converter.NewByteSlicePayloadConverter(),
				pkgdataconverter.NewJSONConverter(),
			))
			env.RegisterWorkflowWithOptions(func(workflow.Context, handlerworkflow.HandlerRequest) error { return nil }, workflow.RegisterOptions{Name: "Handler"})
			if source == "legacy-recovery" {
				env.OnGetVersion("queue-backlog-overflow", workflow.DefaultVersion, 1).Return(workflow.DefaultVersion)
			}
			env.OnActivity(new(activities.Activities).QueueInternalGetQueue, mock.Anything, activities.GetQueueRequest{QueueID: "queue-id"}).
				Return(&app.Queue{MaxDepth: capacity, MaxInFlight: 1}, nil).Once()

			signals := make([]*app.QueueSignal, count)
			for i := range signals {
				signals[i] = &app.QueueSignal{ID: fmt.Sprintf("signal-%02d", i), QueueID: "queue-id"}
				signals[i].Workflow.ID = fmt.Sprintf("handler-%02d", i)
			}
			if source != "enqueue" {
				env.OnActivity(new(activities.Activities).QueueInternalGetQueueSignals, mock.Anything, activities.GetQueueSignalsRequest{QueueID: "queue-id"}).
					Return(signals, nil).Once()
			}

			expectedCount := count
			if source == "legacy-recovery" {
				expectedCount = capacity
			}
			var executed []string
			for _, sig := range signals[:expectedCount] {
				env.OnActivity(new(activities.Activities).QueueInternalGetQueueSignal, mock.Anything, activities.GetQueueSignalRequest{QueueSignalID: sig.ID}).
					Return(sig, nil).Once()
				env.OnActivity(new(handleractivities.Activities).HandlerInternalUpdateWorkflowReady, mock.Anything,
					handleractivities.UpdateWorkflowReadyRequest{QueueID: "queue-id", WorkflowID: sig.Workflow.ID, UpdateID: sig.ID}).
					Return(&handlerworkflow.ReadyResponse{RunID: "handler-run"}, nil).Once()
			}
			env.OnActivity(new(activities.Activities).UpdateQueueSignalRunID, mock.Anything, mock.Anything).Return(nil).Times(expectedCount)
			env.OnActivity(new(statusactivities.Activities).UpdateQueueSignalStatusV2, mock.Anything, mock.Anything).Return(nil).Times(expectedCount)
			env.OnActivity(new(handleractivities.Activities).HandlerInternalUpdateWorkflowValidate, mock.Anything, mock.Anything).
				Return(func(_ context.Context, req handleractivities.UpdateWorkflowValidateRequest) error {
					env.SignalWorkflow(req.Cb.SignalName, callback.Result{Status: "success"})
					return nil
				}).Times(expectedCount)
			env.OnActivity(new(handleractivities.Activities).HandlerInternalUpdateWorkflowExecute, mock.Anything, mock.Anything).
				After(time.Second).
				Return(func(_ context.Context, req handleractivities.UpdateWorkflowExecuteRequest) error {
					executed = append(executed, req.UpdateID)
					env.SignalWorkflow(req.Cb.SignalName, callback.Result{Status: "success"})
					return nil
				}).Times(expectedCount)

			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				q := &queue{queueID: "queue-id", state: &QueueState{}, inFlightSignals: make(map[string]bool)}
				if err := q.setupChannels(ctx); err != nil {
					return err
				}
				if source == "enqueue" {
					q.ready = true
					for _, sig := range signals {
						if err := q.enqueue(ctx, EnqueueHandlerInput{QueueSignalID: sig.ID, WorkflowID: sig.Workflow.ID}); err != nil {
							return err
						}
					}
				} else if err := q.requeueSignals(ctx); err != nil {
					return err
				}
				// Recovery must finish before readiness, even with no consumer yet.
				require.Len(t, q.inFlightSignals, expectedCount)
				q.ready = true
				if source != "legacy-recovery" {
					// A duplicate notification for an overflow item must not start
					// another handler or append a second reference.
					last := signals[count-1]
					require.NoError(t, q.enqueue(ctx, EnqueueHandlerInput{QueueSignalID: last.ID, WorkflowID: last.Workflow.ID}))
				}
				q.lastActivityTime = workflow.Now(ctx).Add(-2 * defaultQueueIdleTimeout)
				require.True(t, q.isIdle(ctx), "inactivity must still trigger draining with a backlog")
				require.Equal(t, source == "legacy-recovery", q.canCompleteIdle(ctx, 0),
					"retained references must prevent idle completion before dispatch")
				q.paused = true
				if err := q.startDispatcher(ctx); err != nil {
					return err
				}
				if err := workflow.Sleep(ctx, 2*defaultQueueIdleTimeout); err != nil {
					return err
				}
				require.Empty(t, executed, "paused queues must not execute")
				require.False(t, q.isIdle(ctx))
				q.paused = false
				if err := workflow.Await(ctx, func() bool {
					require.LessOrEqual(t, q.activeWorkers, 1)
					return len(q.inFlightSignals) == 0
				}); err != nil {
					return err
				}
				require.Empty(t, q.pendingSignals)
				var extra QueueRef
				require.False(t, q.ch.ReceiveAsync(&extra), "duplicate delivery")
				require.Zero(t, q.activeWorkers)
				q.lastActivityTime = workflow.Now(ctx).Add(-2 * defaultQueueIdleTimeout)
				require.True(t, q.isIdle(ctx))
				require.True(t, q.canCompleteIdle(ctx, 0))
				require.False(t, q.canCompleteIdle(ctx, 1), "drained notifications still need recovery")
				q.stopped = true
				return nil
			})
			require.NoError(t, env.GetWorkflowError())
			expected := make([]string, expectedCount)
			for i := range expected {
				expected[i] = fmt.Sprintf("signal-%02d", i)
			}
			require.Equal(t, expected, executed, "signals must execute once, in FIFO order")
			env.AssertExpectations(t)
		})
	}
}

func TestQueueRunInactivity(t *testing.T) {
	for _, scenario := range []string{"empty", "missing-callback", "resume-backlog"} {
		t.Run(scenario, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkflowRunTimeout(time.Minute)
			env.SetDataConverter(converter.NewCompositeDataConverter(
				converter.NewNilPayloadConverter(),
				converter.NewByteSlicePayloadConverter(),
				pkgdataconverter.NewJSONConverter(),
			))
			paused := scenario == "resume-backlog"
			var signals []*app.QueueSignal
			if scenario != "empty" {
				count := 1
				if paused {
					count = 12
				}
				for i := 0; i < count; i++ {
					sig := &app.QueueSignal{ID: fmt.Sprintf("signal-%02d", i), QueueID: "queue-id"}
					sig.Workflow.ID = fmt.Sprintf("handler-%02d", i)
					signals = append(signals, sig)
				}
			}
			env.OnActivity(new(activities.Activities).QueueInternalQueueExists, mock.Anything, mock.Anything).Return(true, nil).Once()
			env.OnActivity(new(activities.Activities).QueueInternalClearRestartHint, mock.Anything, mock.Anything).Return(nil).Once()
			env.OnActivity(new(activities.Activities).QueueInternalCheckRestartHint, mock.Anything, mock.Anything).Return(false, nil).Maybe()
			env.OnActivity(new(activities.Activities).QueueInternalGetQueue, mock.Anything, mock.Anything).
				Return(&app.Queue{MaxDepth: 10, MaxInFlight: 1}, nil)
			env.OnActivity(new(activities.Activities).QueueInternalGetQueueSignals, mock.Anything, mock.Anything).Return(signals, nil).Once()
			env.OnActivity(new(statusactivities.Activities).UpdateQueueStatusV2, mock.Anything, mock.Anything).Return(nil)
			env.OnActivity(new(activities.Activities).QueueInternalUpdateQueueMetadata, mock.Anything, mock.Anything).Return(nil).Maybe()
			if scenario != "empty" {
				env.OnActivity(new(activities.Activities).QueueInternalGetQueueSignal, mock.Anything, activities.GetQueueSignalRequest{QueueSignalID: signals[0].ID}).Return(signals[0], nil).Maybe()
				env.OnActivity(new(statusactivities.Activities).UpdateQueueSignalStatusV2, mock.Anything, mock.Anything).Return(nil).Maybe()
				env.OnActivity(new(activities.Activities).UpdateQueueSignalRunID, mock.Anything, mock.Anything).Return(nil).Maybe()
				env.OnActivity(new(handleractivities.Activities).HandlerInternalUpdateWorkflowReady, mock.Anything, mock.Anything).
					Return(&handlerworkflow.ReadyResponse{RunID: "handler-run"}, nil).Maybe()
				env.OnActivity(new(handleractivities.Activities).HandlerInternalUpdateWorkflowValidate, mock.Anything, mock.Anything).
					Return(func(_ context.Context, req handleractivities.UpdateWorkflowValidateRequest) error {
						env.SignalWorkflow(req.Cb.SignalName, callback.Result{Status: "success"})
						return nil
					}).Maybe()
				// The execute update is accepted, but the handler never calls back.
				execute := env.OnActivity(new(handleractivities.Activities).HandlerInternalUpdateWorkflowExecute, mock.Anything, mock.Anything).Return(nil)
				if scenario == "missing-callback" {
					execute.Once()
				} else {
					execute.Maybe()
				}
			}

			var elapsed time.Duration
			env.ExecuteWorkflow(func(ctx workflow.Context) (bool, error) {
				start := workflow.Now(ctx)
				q := &queue{
					queueID: "queue-id", state: &QueueState{}, paused: paused,
					inFlightSignals: make(map[string]bool), lastActivityTime: start,
					cfg: &internal.Config{
						QueueIdleTimeout: 10 * time.Second, QueueDrainTimeout: 5 * time.Second,
						QueueContinueAsNewHintPeriod: time.Second,
					},
				}
				if paused {
					workflow.Go(ctx, func(ctx workflow.Context) {
						if err := workflow.Sleep(ctx, 20*time.Second); err == nil {
							// Retain more than channel capacity through a pause
							// longer than the idle timeout before resuming.
							q.paused = false
						}
					})
				}
				finished, err := q.run(ctx)
				elapsed = workflow.Now(ctx).Sub(start)
				return finished, err
			})
			require.NoError(t, env.GetWorkflowError())
			var finished bool
			require.NoError(t, env.GetWorkflowResult(&finished))
			require.Equal(t, scenario == "empty", finished,
				"only an empty queue may complete idle; retained work needs continue-as-new")
			require.Less(t, elapsed, time.Minute, "inactivity recovery must not wait for the handler timeout")
			if scenario == "missing-callback" {
				require.GreaterOrEqual(t, elapsed, 15*time.Second, "allow both idle and drain timeouts")
			}
			env.AssertExpectations(t)
		})
	}
}
