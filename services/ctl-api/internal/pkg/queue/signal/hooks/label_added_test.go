package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestLabelAddedWebhookData(t *testing.T) {
	installID := "ins-test"
	event := signal.SignalPhaseEvent{
		SignalType: signalTypeLabelAdded,
		Phase:      signal.SignalPhaseExecute,
		OrgID:      "org-test",
		OwnerID:    installID,
		OwnerType:  "installs",
		InstallID:  &installID,
		Metadata: map[string]any{
			"label_name": "environment",
		},
	}

	hook := &WebhookSignalLifecycleHook{}
	data, ok := hook.buildEventData(context.Background(), event, &signal.SignalPhaseOutcome{
		Status: signal.SignalStatusSuccess,
	})

	require.True(t, ok)
	require.Equal(t, kindLabelAdded, data.Kind)
	require.Equal(t, transitionSucceeded, data.Transition)
	require.Equal(t, "environment", data.Metadata["label_name"])
	require.Equal(t, installID, EventTargetsFromEvent(context.Background(), nil, event, data).InstallID)
}
