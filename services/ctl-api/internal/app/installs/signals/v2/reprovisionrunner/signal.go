package reprovisionrunner

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/reprovisionserviceaccount"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "reprovision-runner"

type Signal struct {
	signal.LifecycleBase

	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID:    &s.InstallID,
		Operation:    "runner-reprovision",
		WorkflowID:   s.LifecycleWorkflowID,
		WorkflowType: s.LifecycleWorkflowType,
	}
}

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
	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	// Enqueue reprovision-service-account signal to the runner's queue
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   install.RunnerID,
		OwnerType: "runners",
		Signal: &runnersignals.Signal{
			RunnerID: install.RunnerID,
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue reprovision service account signal to runner")
	}

	return nil
}
