package driftchecksandbox

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const SignalType signal.SignalType = "drift_check_sandbox"

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	wkflw, err := activities.AwaitCreateWorkflow(ctx, activities.CreateWorkflowRequest{
		InstallID:    s.InstallID,
		WorkflowType: app.WorkflowTypeDriftRunReprovisionSandbox,
		Metadata:     map[string]string{},
		PlanOnly:     true,
	})
	if err != nil {
		return fmt.Errorf("unable to create workflow: %w", err)
	}

	return signalsactivities.AwaitPkgSignalsSendInstallsSignal(ctx, &signalsactivities.SendSignalRequest[*installsignals.Signal]{
		ID: s.InstallID,
		Signal: &installsignals.Signal{
			Type:              installsignals.OperationExecuteFlow,
			InstallWorkflowID: wkflw.ID,
		},
	})
}
