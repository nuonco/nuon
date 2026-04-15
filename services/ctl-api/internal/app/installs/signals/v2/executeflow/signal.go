package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "execute-flow"

type Signal struct {
	signal.Hooks
	InstallWorkflowID string `json:"install_workflow_id"`

	// installID is resolved from the workflow's OwnerID during Validate
	installID string
}

var _ qsignal.Signal = (*Signal)(nil)

func (s *Signal) Type() qsignal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallWorkflowID == "" {
		return errors.New("install_workflow_id is required")
	}

	// Resolve install ID from the workflow's OwnerID
	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.InstallWorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow")
	}
	if flw.OwnerID == "" || flw.OwnerType != "installs" {
		return errors.New("workflow does not belong to an install")
	}
	s.installID = flw.OwnerID

	// Validate install exists
	_, err = activities.AwaitGetByInstallID(ctx, s.installID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return s.executeFlow(ctx)
}
