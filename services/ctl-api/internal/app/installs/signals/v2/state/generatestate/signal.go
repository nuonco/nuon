package generatestate

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/stateregenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "generate-state"

type Signal struct {
	InstallID string
}

var (
	_ signal.Signal              = &Signal{}
	_ signal.SignalWithAutoRetry = (*Signal)(nil)
)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install id is required")
	}

	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		QueueName: installshelpers.InstallStateManagerQueueName,
		Signal: &stateregenerate.Signal{
			InstallID:        s.InstallID,
			Targets:          pkgstate.AllPartialTargets(),
			ForceAll:         true,
			TriggeredByID:    s.InstallID,
			TriggeredByType:  "installs",
			StateGeneratedBy: app.InstallStateGenerateSourceStateManager,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to force-regenerate state")
	}
	return nil
}
