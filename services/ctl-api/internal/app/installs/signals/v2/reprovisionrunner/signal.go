package reprovisionrunner

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "reprovision-runner"

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

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get the install to find the runner ID (copied from worker/reprovision_runner.go)
	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// TODO: Send signal to runner namespace to reprovision the service account
	// Original code: w.evClient.Send(ctx, install.RunnerID, &runnersignals.Signal{Type: runnersignals.OperationReprovisionServiceAccount})
	//
	// This is deprecated and needs to be adapted for the queue system.
	// Cross-namespace signal sending needs to be implemented as part of the queue infrastructure.
	_ = install.RunnerID // suppress unused variable warning

	return nil
}
