package builds

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-builds"

type Signal struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	RunID       string `json:"run_id" validate:"required"`

	// FlowID and StepID are injected by the flow engine via SignalWithStepContext.
	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Use playground validator for struct tag validation
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()

	l := workflow.GetLogger(cancelCtx)

	inflight, err := activities.AwaitGetInflightBuildQueueSignalsByRunID(cancelCtx, s.RunID)
	if err != nil {
		l.Warn("failed to get inflight build queue signals for cancel", "error", err)
		return nil
	}

	for _, qs := range inflight {
		if _, err := queueclient.AwaitCancelSignal(cancelCtx, qs.QueueSignalID); err != nil {
			l.Warn("failed to cancel build signal",
				"queue_signal_id", qs.QueueSignalID,
				"build_id", qs.BuildID,
				"error", err)
		}
	}

	return nil
}
