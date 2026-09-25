package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

func TestResolveFlowCompletionOutcome(t *testing.T) {
	tests := map[string]struct {
		rowStatus    *app.CompositeStatus
		lookupErr    error
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
		"status lookup failure fails closed": {
			rowStatus: &app.CompositeStatus{Status: app.StatusSuccess},
			lookupErr: errors.New("db down"),
			event: signal.SignalPhaseEvent{
				SignalType: signalTypeExecuteWorkflow,
				Phase:      signal.SignalPhaseExecute,
				WorkflowID: "wfl_1",
			},
			outcome:     signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			wantErr:     true,
			wantOutcome: signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
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
			lookup := func(_ context.Context, workflowID string) (app.CompositeStatus, error) {
				if tt.lookupErr != nil {
					return app.CompositeStatus{}, tt.lookupErr
				}
				if tt.rowStatus == nil || workflowID != "wfl_1" {
					return app.CompositeStatus{}, gorm.ErrRecordNotFound
				}
				return *tt.rowStatus, nil
			}

			outcome := tt.outcome
			suppress, err := resolveFlowCompletionOutcome(context.Background(), lookup, tt.event, &outcome)
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

// TestNilDBSkipsLookup proves hooks without a DB publish the transport outcome
// unchanged instead of failing closed.
func TestNilDBSkipsLookup(t *testing.T) {
	event := signal.SignalPhaseEvent{
		SignalType: signalTypeExecuteWorkflow,
		Phase:      signal.SignalPhaseExecute,
		WorkflowID: "wfl_1",
	}
	outcome := signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}
	suppress, err := resolveFlowCompletionOutcome(context.Background(), workflowStatusFromDB(nil), event, &outcome)
	require.NoError(t, err)
	assert.False(t, suppress)
	assert.Equal(t, signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}, outcome)
}

// TestFlowCompletionStatusLookupFailureFailsClosed drives the hooks against a
// PostgreSQL DSN nothing listens on, so the status read fails after retries
// and AfterPhase must refuse to publish.
func TestFlowCompletionStatusLookupFailureFailsClosed(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=127.0.0.1 port=1 user=unused dbname=unused sslmode=disable connect_timeout=1",
	}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

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
