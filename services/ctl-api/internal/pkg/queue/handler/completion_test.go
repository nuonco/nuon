package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type callbackGateSignal struct {
	workflowID string
}

func (s *callbackGateSignal) Type() signal.SignalType               { return "callback-gate-test" }
func (s *callbackGateSignal) Validate(workflow.Context) error       { return nil }
func (s *callbackGateSignal) Execute(workflow.Context) error        { return nil }
func (s *callbackGateSignal) CompletionCallbacksWorkflowID() string { return s.workflowID }

type ordinaryCallbackSignal struct{}

func (s *ordinaryCallbackSignal) Type() signal.SignalType         { return "ordinary-callback-test" }
func (s *ordinaryCallbackSignal) Validate(workflow.Context) error { return nil }
func (s *ordinaryCallbackSignal) Execute(workflow.Context) error  { return nil }

func TestCompletionCallbacksWorkflowID(t *testing.T) {
	assert.Equal(t, "wfl_1", completionCallbacksWorkflowID(&callbackGateSignal{workflowID: "wfl_1"}))
	assert.Empty(t, completionCallbacksWorkflowID(&callbackGateSignal{}))
	assert.Empty(t, completionCallbacksWorkflowID(&ordinaryCallbackSignal{}))
}

func TestSendCompletionCallbacksGate(t *testing.T) {
	tests := map[string]struct {
		sig       signal.Signal
		hold      bool
		gateErr   error
		wantCalls []string
	}{
		"resident parked workflow holds callback": {
			sig:       &callbackGateSignal{workflowID: "wfl_1"},
			hold:      true,
			wantCalls: []string{"gate"},
		},
		"resident terminal workflow sends callback": {
			sig:       &callbackGateSignal{workflowID: "wfl_1"},
			wantCalls: []string{"gate", "reload", "send"},
		},
		"gate lookup failure preserves callback delivery": {
			sig:       &callbackGateSignal{workflowID: "wfl_1"},
			gateErr:   errors.New("database unavailable"),
			wantCalls: []string{"gate", "reload", "send"},
		},
		"ordinary signal does not use gate": {
			sig:       &ordinaryCallbackSignal{},
			wantCalls: []string{"reload", "send"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.RegisterActivityWithOptions(func(context.Context, any) error { return nil }, activity.RegisterOptions{Name: "SendSignal"})
			calls := make([]string, 0, len(tt.wantCalls))

			if completionCallbacksWorkflowID(tt.sig) != "" {
				env.OnActivity((*activities.Activities).HoldCompletionCallbacks, mock.Anything, mock.Anything).
					Run(func(mock.Arguments) { calls = append(calls, "gate") }).
					Return(tt.hold, tt.gateErr).
					Once()
			}

			if !tt.hold {
				env.OnActivity((*activities.Activities).QueueInternalGetQueueSignal, mock.Anything, mock.Anything).
					Run(func(mock.Arguments) { calls = append(calls, "reload") }).
					Return(&app.QueueSignal{
						Callbacks: callback.Refs{{
							WorkflowID: "parent-workflow",
							SignalName: "complete",
						}},
					}, nil).
					Once()
				env.OnActivity("SendSignal", mock.Anything, mock.Anything).
					Run(func(mock.Arguments) { calls = append(calls, "send") }).
					Return(nil).
					Once()
			}

			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				h := &handler{
					queueSignalID:  "queue-signal",
					sig:            tt.sig,
					finishedStatus: app.StatusSuccess,
				}
				h.sendCompletionCallbacks(ctx)
				return nil
			})

			require.True(t, env.IsWorkflowCompleted())
			require.NoError(t, env.GetWorkflowError())
			assert.Equal(t, tt.wantCalls, calls)
			env.AssertExpectations(t)
		})
	}
}
