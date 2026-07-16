package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/slackrender"
)

func awaitingRetryEvent() signal.SignalPhaseEvent {
	return signal.SignalPhaseEvent{
		QueueSignalID: "sig_1",
		SignalType:    signalTypeWorkflowStepAwaitingRetry,
		Phase:         signal.SignalPhaseExecute,
		OrgID:         "org_1",
		OrgName:       "acme",
		WorkflowID:    "wfl_1",
		WorkflowType:  "deploy_components",
		StepID:        "iws_1",
		OwnerID:       "ins_1",
		OwnerType:     "installs",
		Metadata: map[string]any{
			"awaiting_retry":         true,
			"manual_action_required": true,
			"terminal":               false,
			"error":                  "terraform apply failed",
			"step_name":              "deploy my-component",
			"retry_index":            2,
			"max_retries":            5,
		},
	}
}

func TestAwaitingRetryBuildEventData(t *testing.T) {
	h := &WebhookSignalLifecycleHook{l: zap.NewNop()}
	ctx := context.Background()

	t.Run("successful carrier projects failed non-terminal step event", func(t *testing.T) {
		event := awaitingRetryEvent()
		outcome := &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}

		data, ok := h.buildEventData(ctx, event, outcome)
		require.True(t, ok)

		assert.Equal(t, kindWorkflowStep, data.Kind)
		assert.Equal(t, transitionAwaitingRetry, data.Transition)
		assert.Equal(t, "org_1", data.OrgID)
		assert.Equal(t, "wfl_1", data.Workflow.ID)
		assert.Equal(t, "ins_1", data.Workflow.OwnerID)
		assert.Equal(t, "installs", data.Workflow.OwnerType)

		require.NotNil(t, data.Step)
		assert.Equal(t, "iws_1", data.Step.ID)

		require.NotNil(t, data.Outcome)
		assert.Equal(t, statusFailed, data.Outcome.Status)
		assert.Equal(t, "terraform apply failed", data.Outcome.Error)

		require.NotNil(t, data.Metadata)
		assert.Equal(t, true, data.Metadata["awaiting_retry"])
		assert.Equal(t, true, data.Metadata["manual_action_required"])
		assert.Equal(t, false, data.Metadata["terminal"])
		assert.Equal(t, 2, data.Metadata["retry_index"])
		assert.Equal(t, 5, data.Metadata["max_retries"])
	})

	t.Run("before-phase (nil outcome) produces no event", func(t *testing.T) {
		_, ok := h.buildEventData(ctx, awaitingRetryEvent(), nil)
		assert.False(t, ok)
	})

	t.Run("failed carrier produces no event", func(t *testing.T) {
		_, ok := h.buildEventData(ctx, awaitingRetryEvent(), &signal.SignalPhaseOutcome{
			Status:     signal.SignalStatusError,
			ErrMessage: "notification carrier broke",
		})
		assert.False(t, ok)
	})

	t.Run("cancelled carrier produces no event", func(t *testing.T) {
		_, ok := h.buildEventData(ctx, awaitingRetryEvent(), &signal.SignalPhaseOutcome{
			Status: signal.SignalStatusCancelled,
		})
		assert.False(t, ok)
	})

	t.Run("missing error metadata keeps carrier error empty but status failed", func(t *testing.T) {
		event := awaitingRetryEvent()
		delete(event.Metadata, "error")

		data, ok := h.buildEventData(ctx, event, &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess})
		require.True(t, ok)
		require.NotNil(t, data.Outcome)
		assert.Equal(t, statusFailed, data.Outcome.Status)
		assert.Empty(t, data.Outcome.Error)
	})
}

func TestAwaitingRetryStartedEventSuppression(t *testing.T) {
	assert.True(t, suppressesStartedEvent(signalTypeWorkflowStepAwaitingRetry))
	assert.False(t, isNotificationOnlySignalType(signalTypeWorkflowStepAwaitingRetry))
}

func TestAwaitingRetrySlackRendering(t *testing.T) {
	event := awaitingRetryEvent()
	data, ok := (&WebhookSignalLifecycleHook{l: zap.NewNop()}).
		buildEventData(context.Background(), event, &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess})
	require.True(t, ok)

	rendered := buildRenderEvent(data)
	assert.Equal(t, slackrender.TransitionAwaitingRetry, rendered.event.Transition)
	require.NotNil(t, rendered.event.Outcome)
	assert.Equal(t, statusFailed, rendered.event.Outcome.Status)
	assert.Equal(t, "terraform apply failed", rendered.event.Outcome.Error)
}
