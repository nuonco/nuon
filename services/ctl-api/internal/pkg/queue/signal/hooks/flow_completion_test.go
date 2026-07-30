package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

func TestSuppressParkedFlowCompletion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&workflowStatusRow{}))

	ctx := context.Background()
	event := signal.SignalPhaseEvent{
		SignalType: signalTypeExecuteWorkflow,
		Phase:      signal.SignalPhaseExecute,
		WorkflowID: "wfl_1",
		OwnerType:  "installs",
	}
	outcome := signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}

	require.NoError(t, db.Create(&workflowStatusRow{
		ID:        "wfl_1",
		OwnerType: "installs",
		Status: app.CompositeStatus{
			Status: app.StatusFailedPendingRetry,
		},
	}).Error)

	suppressed, err := suppressParkedFlowCompletion(ctx, db, event, outcome)
	require.NoError(t, err)
	assert.True(t, suppressed)

	require.NoError(t, db.Save(&workflowStatusRow{
		ID:        event.WorkflowID,
		OwnerType: "installs",
		Status: app.CompositeStatus{
			Status: app.StatusSuccess,
		},
	}).Error)

	suppressed, err = suppressParkedFlowCompletion(ctx, db, event, outcome)
	require.NoError(t, err)
	assert.False(t, suppressed)

	require.NoError(t, db.Save(&workflowStatusRow{
		ID:        event.WorkflowID,
		OwnerType: "apps",
		Status: app.CompositeStatus{
			Status: app.StatusFailedPendingRetry,
		},
	}).Error)

	suppressed, err = suppressParkedFlowCompletion(ctx, db, event, outcome)
	require.NoError(t, err)
	assert.False(t, suppressed)

	event.SignalType = signalTypeExecuteWorkflowStep
	suppressed, err = suppressParkedFlowCompletion(ctx, db, event, outcome)
	require.NoError(t, err)
	assert.False(t, suppressed)

	event.SignalType = signalTypeExecuteWorkflow
	event.WorkflowID = "missing"
	outcome.Status = signal.SignalStatusError
	suppressed, err = suppressParkedFlowCompletion(ctx, db, event, outcome)
	require.NoError(t, err)
	assert.False(t, suppressed)
}

func TestFlowCompletionStatusLookupFailureDoesNotSuppressHooks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&workflowStatusRow{}))

	event := signal.SignalPhaseEvent{
		SignalType: signalTypeExecuteWorkflow,
		Phase:      signal.SignalPhaseExecute,
		WorkflowID: "missing",
	}
	outcome := signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}

	webhookHook := &WebhookSignalLifecycleHook{l: zap.NewNop(), db: db}
	require.NoError(t, webhookHook.AfterPhase(context.Background(), event, outcome))

	slackHook := &SlackSignalLifecycleHook{
		l:           zap.NewNop(),
		db:          db,
		slackClient: &slackclient.Client{},
		enricher:    webhookHook,
	}
	require.NoError(t, slackHook.AfterPhase(context.Background(), event, outcome))
}
