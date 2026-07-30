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

func TestResolveFlowCompletionOutcome(t *testing.T) {
	tests := map[string]struct {
		rowStatus    *app.CompositeStatus
		event        signal.SignalPhaseEvent
		outcome      signal.SignalPhaseOutcome
		wantSuppress bool
		wantErr      bool
		wantOutcome  signal.SignalPhaseOutcome
	}{
		"parked workflow suppresses completion": {
			rowStatus: &app.CompositeStatus{Status: app.StatusFailedPendingRetry},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome:      signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantSuppress: true,
			wantOutcome:  signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
		},
		"parked non-install workflow suppresses completion": {
			rowStatus: &app.CompositeStatus{Status: app.StatusFailedPendingRetry},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
				OwnerType:  "apps",
			},
			outcome:      signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantSuppress: true,
			wantOutcome:  signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
		},
		"successful workflow publishes unchanged": {
			rowStatus: &app.CompositeStatus{Status: app.StatusSuccess},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome:     signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantOutcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
		},
		"errored workflow rewrites transport success to error": {
			rowStatus: &app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "step deploy-app failed",
			},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantOutcome: signal.SignalPhaseOutcome{
				Status:     signal.SignalStatusError,
				ErrMessage: "step deploy-app failed",
			},
		},
		"errored workflow with empty description gets default error text": {
			rowStatus: &app.CompositeStatus{Status: app.StatusError},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantOutcome: signal.SignalPhaseOutcome{
				Status:     signal.SignalStatusError,
				ErrMessage: "workflow failed",
			},
		},
		"cancelled workflow with empty description gets default cancel text": {
			rowStatus: &app.CompositeStatus{Status: app.StatusCancelled},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantOutcome: signal.SignalPhaseOutcome{
				Status:     signal.SignalStatusCancelled,
				ErrMessage: "workflow cancelled",
			},
		},
		"cancelled workflow rewrites transport success to cancelled": {
			rowStatus: &app.CompositeStatus{
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "cancelled by user",
			},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantOutcome: signal.SignalPhaseOutcome{
				Status:     signal.SignalStatusCancelled,
				ErrMessage: "cancelled by user",
			},
		},
		"step signal type is untouched": {
			rowStatus: &app.CompositeStatus{Status: app.StatusFailedPendingRetry},
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflowStep,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome:     signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantOutcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
		},
		"transport error is preserved without lookup": {
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_missing",
			},
			outcome: signal.SignalPhaseOutcome{
				Status:     signal.SignalStatusError,
				ErrMessage: "queue transport failed",
			},
			wantOutcome: signal.SignalPhaseOutcome{
				Status:     signal.SignalStatusError,
				ErrMessage: "queue transport failed",
			},
		},
		"missing workflow row fails closed": {
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_missing",
			},
			outcome:     signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantErr:     true,
			wantOutcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, db.AutoMigrate(&workflowStatusRow{}))

			if tt.rowStatus != nil {
				require.NoError(t, db.Create(&workflowStatusRow{
					ID:     "wfl_1",
					Status: *tt.rowStatus,
				}).Error)
			}

			outcome := tt.outcome
			suppress, err := resolveFlowCompletionOutcome(context.Background(), db, tt.event, &outcome)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantSuppress, suppress)
			assert.Equal(t, tt.wantOutcome, outcome)
		})
	}
}

func TestFlowCompletionStatusLookupFailureFailsClosed(t *testing.T) {
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
	require.Error(t, webhookHook.AfterPhase(context.Background(), event, outcome))

	slackHook := &SlackSignalLifecycleHook{
		l:           zap.NewNop(),
		db:          db,
		slackClient: &slackclient.Client{},
		enricher:    webhookHook,
	}
	require.Error(t, slackHook.AfterPhase(context.Background(), event, outcome))
}
