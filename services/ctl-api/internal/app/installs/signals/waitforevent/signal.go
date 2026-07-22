package waitforevent

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "wait-for-event"
const updateName = "event-received"

type Signal struct {
	InstallID      string              `json:"install_id"`
	WorkflowID     string              `json:"workflow_id"`
	WorkflowStepID string              `json:"workflow_step_id"`
	QueueSignalID  string              `json:"queue_signal_id"`
	TriggerID      string              `json:"trigger_id"`
	EventTypes     []string            `json:"event_types"`
	Filters        []app.TriggerFilter `json:"filters,omitempty"`
	WaitTimeout    time.Duration       `json:"timeout,omitempty"`
	matched        bool
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithParams = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithUpdateHandlers = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithTimeout = (*Signal)(nil)
var _ signal.SignalWithUnboundedTimeout = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType     { return SignalType }
func (s *Signal) Timeout() time.Duration      { return s.WaitTimeout }
func (s *Signal) UnboundedTimeout() bool      { return s.WaitTimeout <= 0 }
func (s *Signal) WithParams(p *signal.Params) { s.QueueSignalID = p.QueueSignalID }
func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID, s.WorkflowID = stepID, flowID
}
func (s *Signal) Validate(workflow.Context) error {
	if s.InstallID == "" || s.TriggerID == "" || len(s.EventTypes) == 0 {
		return fmt.Errorf("install_id, trigger_id, and event_types are required")
	}
	return nil
}
func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	return workflow.SetUpdateHandler(ctx, updateName, func(workflow.Context) error { s.matched = true; return nil })
}
func (s *Signal) Execute(ctx workflow.Context) error {
	waiter, err := activities.AwaitRegisterEventRunbookWaiter(ctx, activities.RegisterEventRunbookWaiterRequest{InstallID: s.InstallID, WorkflowID: s.WorkflowID, WorkflowStepID: s.WorkflowStepID, QueueSignalID: s.QueueSignalID, TriggerID: s.TriggerID, EventTypes: s.EventTypes, Filters: s.Filters})
	if err != nil {
		return err
	}
	if waiter.Status == app.EventRunbookWaiterStatusMatched {
		return nil
	}
	if s.WaitTimeout <= 0 {
		if err := workflow.Await(ctx, func() bool { return s.matched }); err != nil {
			return err
		}
		return nil
	}
	if matched, err := workflow.AwaitWithTimeout(ctx, s.WaitTimeout, func() bool { return s.matched }); err == nil && matched {
		return nil
	}
	status, err := activities.AwaitFinishEventRunbookWaiter(ctx, activities.FinishEventRunbookWaiterRequest{WorkflowStepID: s.WorkflowStepID, InstallID: s.InstallID, Status: app.EventRunbookWaiterStatusExpired})
	if err != nil {
		return err
	}
	if status == app.EventRunbookWaiterStatusMatched {
		return nil
	}
	return fmt.Errorf("timed out waiting for event")
}
func (s *Signal) Cancel(ctx workflow.Context) error {
	dctx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	_, err := activities.AwaitFinishEventRunbookWaiter(dctx, activities.FinishEventRunbookWaiterRequest{WorkflowStepID: s.WorkflowStepID, InstallID: s.InstallID, Status: app.EventRunbookWaiterStatusCancelled})
	return err
}
