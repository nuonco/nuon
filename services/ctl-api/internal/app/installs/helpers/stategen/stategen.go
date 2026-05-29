package stategen

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/statepartialgenerate"
	workerstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Request struct {
	StateGenV2      bool
	InstallID       string
	Targets         []state.PartialTarget
	AllTargets      bool
	ForceAll        bool
	TriggeredByID   string
	TriggeredByType string
}

// HintOrGenerate hints the state manager via state-partial-generate when StateGenV2
// is set and the install has a state-manager queue; otherwise (and on missing-queue
// fallback) it runs legacy in-band generation.
func HintOrGenerate(ctx workflow.Context, req Request) error {
	if req.StateGenV2 {
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
			if !generics.IsGormErrRecordNotFound(err) {
				return errors.Wrap(err, "unable to hint state manager")
			}
			// state-manager queue missing — fall through to legacy generation
		} else {
			if _, err := callback.Await(ctx, cb); err != nil {
				return errors.Wrap(err, "unable to await state generation")
			}
			return nil
		}
	}

	if _, err := workerstate.AwaitGenerateState(ctx, &workerstate.GenerateStateRequest{
		InstallID:       req.InstallID,
		TriggeredByID:   req.TriggeredByID,
		TriggeredByType: req.TriggeredByType,
	}); err != nil {
		return errors.Wrap(err, "unable to generate state")
	}
	return nil
}
