package generatestate

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"

	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "generate-full-state"

// This signal force generates all state parts and calls the state-regenerate signal with
// forceAll: true and targets: all-partials
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
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		QueueName: installshelpers.InstallStateManagerQueueName,
		Signal: &statepartialgenerate.Signal{
			InstallID:       s.InstallID,
			AllTargets:      true,
			TriggeredByID:   s.InstallID,
			TriggeredByType: "installs",
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to force-regenerate state")
	}
	if _, err := queueclient.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID); err != nil {
		return errors.Wrap(err, "unable to await state generation")
	}
	return nil
}
