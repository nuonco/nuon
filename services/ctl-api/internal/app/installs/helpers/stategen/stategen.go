package stategen

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Request struct {
	InstallID       string
	Targets         []state.PartialTarget
	AllTargets      bool
	ForceAll        bool
	TriggeredByID   string
	TriggeredByType string
}

// HintOrGenerate hints the state manager via state-partial-generate on the
// install's state-manager queue and awaits the regeneration callback.
func HintOrGenerate(ctx workflow.Context, req Request) error {
	cb := callback.New(ctx, req.InstallID)
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         req.InstallID,
		OwnerType:       "installs",
		SignalOwnerID:   req.InstallID,
		SignalOwnerType: "installs",
		QueueName:       helpers.InstallStateManagerQueueName,
		Signal: &statepartialgenerate.Signal{
			InstallID:       req.InstallID,
			Targets:         req.Targets,
			AllTargets:      req.AllTargets,
			ForceAll:        req.ForceAll,
			TriggeredByID:   req.TriggeredByID,
			TriggeredByType: req.TriggeredByType,
		},
		Callback: cb,
	})
	if err != nil {
		return errors.Wrap(err, "unable to hint state manager")
	}
	if _, err := callback.AwaitWithTimeout(ctx, cb, callback.ShortTimeout); err != nil {
		return errors.Wrap(err, "unable to await state generation")
	}
	return nil
}
