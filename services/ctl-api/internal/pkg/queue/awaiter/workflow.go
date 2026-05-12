package awaiter

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

const (
	// continueAsNewInterval controls how often the workflow continues-as-new
	// to keep its history small.
	continueAsNewInterval = 15 * time.Minute
)

type Params struct {
	fx.In
}

func NewWorkflows(params Params) (*Workflows, error) {
	return &Workflows{}, nil
}

type Workflows struct{}

func (w *Workflows) All() []any {
	return []any{
		w.AwaitSignalWorkflow,
	}
}

type AwaitSignalWorkflowRequest struct {
	QueueSignalID string `validate:"required"`
	Timeout       time.Duration
}

// @temporal-gen-v2 workflow
// @task-queue "api"
// @id-template await-signal-{{.Req.QueueSignalID}}
// @memo type await-signal
func (w *Workflows) AwaitSignalWorkflow(ctx workflow.Context, req AwaitSignalWorkflowRequest) (*handler.FinishedResponse, error) {
	// Check DB for terminal status before waiting. This catches the case
	// where the handler finished before we started or during a continue-as-new.
	qs, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, req.QueueSignalID)
	if err != nil {
		return nil, err
	}

	if isTerminalStatus(qs.Status.Status) {
		return terminalResponse(qs.Status.Status, qs.Status.StatusHumanDescription)
	}

	// Listen for the handler's "done" signal with a continue-as-new timer.
	doneCh := workflow.GetSignalChannel(ctx, handler.DoneSignalName)
	timerCtx, timerCancel := workflow.WithCancel(ctx)
	timerFuture := workflow.NewTimer(timerCtx, continueAsNewInterval)

	sel := workflow.NewSelector(ctx)

	var result *handler.FinishedResponse
	var done bool

	sel.AddReceive(doneCh, func(ch workflow.ReceiveChannel, more bool) {
		var resp handler.FinishedResponse
		ch.Receive(ctx, &resp)
		result = &resp
		done = true
		timerCancel()
	})

	sel.AddFuture(timerFuture, func(f workflow.Future) {
		// Timer fired — continue-as-new to keep history small.
		_ = f.Get(timerCtx, nil)
	})

	sel.Select(ctx)

	if done {
		if result.Status == app.StatusError {
			msg := result.StatusDescription
			if msg == "" {
				msg = fmt.Sprintf("signal execution failed with status: %s", result.Status)
			}
			return nil, temporal.NewNonRetryableApplicationError(msg, "SIGNAL_FAILED", nil)
		}
		return result, nil
	}

	// Timer fired: continue-as-new.
	return nil, workflow.NewContinueAsNewError(ctx, w.AwaitSignalWorkflow, req)
}

func isTerminalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled:
		return true
	default:
		return false
	}
}

func terminalResponse(status app.Status, description string) (*handler.FinishedResponse, error) {
	if status == app.StatusError {
		msg := description
		if msg == "" {
			msg = fmt.Sprintf("signal execution failed with status: %s", status)
		}
		return nil, temporal.NewNonRetryableApplicationError(msg, "SIGNAL_FAILED", nil)
	}
	return &handler.FinishedResponse{Status: status, StatusDescription: description}, nil
}
