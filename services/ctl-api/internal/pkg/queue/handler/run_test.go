package handler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type drainTestSignal struct{}

const drainTestSignalType signal.SignalType = "handler-drain-test"

func (s *drainTestSignal) Type() signal.SignalType         { return drainTestSignalType }
func (s *drainTestSignal) Validate(workflow.Context) error { return nil }
func (s *drainTestSignal) Execute(workflow.Context) error  { return nil }

func init() {
	catalog.Register(drainTestSignalType, func() signal.Signal { return &drainTestSignal{} })
}

// A signal cancelled via the DB fallback before its handler workflow existed
// starts exactly one run, which enters the terminal-drain path. That run must
// notify the parents (no earlier run could have) and answer the dispatcher's
// validate with cancelled instead of stamping in-progress over it.
func TestRunCancelledBeforeStartNotifiesParentsAndRejectsValidate(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetDataConverter(converter.NewCompositeDataConverter(
		signaldb.NewPayloadConverter(),
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		converter.NewJSONPayloadConverter(),
	))
	env.RegisterActivityWithOptions(func(context.Context, any) error { return nil }, activity.RegisterOptions{Name: "SendSignal"})

	qs := &app.QueueSignal{
		ID:             "queue-signal",
		Status:         app.CompositeStatus{Status: app.StatusCancelled},
		ExecutionCount: 0,
		Callbacks:      callback.Refs{{WorkflowID: "parent-workflow", SignalName: "complete"}},
	}
	env.OnActivity((*activities.Activities).QueueInternalGetQueueSignal, mock.Anything, mock.Anything).Return(qs, nil)
	env.OnActivity((*activities.Activities).QueueInternalGetQueueSignalSignal, mock.Anything, mock.Anything).Return(&drainTestSignal{}, nil)

	var sent []map[string]any
	env.OnActivity("SendSignal", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if req, ok := args.Get(1).(map[string]any); ok {
				payload, _ := req["payload"].(map[string]any)
				sent = append(sent, payload)
			}
		}).
		Return(nil)

	var statusWrites []app.Status
	env.OnActivity((*statusactivities.Activities).UpdateQueueSignalStatusV2, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if req, ok := args.Get(1).(statusactivities.UpdateQueueSignalStatusV2Request); ok {
				statusWrites = append(statusWrites, req.Status)
			}
		}).
		Return(nil).
		Maybe()

	var validateErr error
	validated := false
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(ValidateUpdateName, "validate-1", &testsuite.TestUpdateCallback{
			OnAccept:   func() {},
			OnReject:   func(err error) { validateErr, validated = err, true },
			OnComplete: func(_ any, err error) { validateErr, validated = err, true },
		}, callback.Ref{})
	}, time.Second)

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		h := &handler{queueSignalID: qs.ID}
		_, err := h.run(ctx)
		return err
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	require.Len(t, sent, 1, "parent callback must be sent exactly once")
	assert.Equal(t, string(app.StatusCancelled), sent[0]["status"])

	require.True(t, validated, "validate update was not served")
	require.Error(t, validateErr)
	assert.Contains(t, validateErr.Error(), "signal was canceled")

	assert.Empty(t, statusWrites, "cancelled must not be overwritten by dispatch stamps")
	env.AssertExpectations(t)
}
