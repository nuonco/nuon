package generateinstallstackversion

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/provisionserviceaccount"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// prepare runs the install-state generation and runner service-account work that
// provision used to dispatch as two separate steps ahead of this one. Each step
// cost ~5s of nested queue-signal overhead plus a group boundary for a fraction
// of that in real work, so provision folds them in here.
//
// Only provision sets PrepareStateAndRunner. Every other caller leaves it unset
// and this is a no-op.
func (s *Signal) prepare(ctx workflow.Context, install *app.Install) error {
	// Only the enqueue overlaps. The runner provisions its service account on its
	// own queue, so starting it here instead of in a dedicated step gets that
	// work moving while install state generates.
	saDone := workflow.NewChannel(ctx)
	var saErr error
	workflow.Go(ctx, func(gCtx workflow.Context) {
		saErr = enqueueRunnerServiceAccount(gCtx, install.RunnerID)
		saDone.Send(gCtx, true)
	})

	stateErr := regenerateState(ctx, regenerateStateRequest{
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintInstallCreated, ""),
		TriggeredByID:   install.ID,
		TriggeredByType: "installs",
		MetricsWriter:   s.metrics,
	})

	saDone.Receive(ctx, nil)

	if stateErr != nil {
		return stateErr
	}
	if saErr != nil {
		return saErr
	}

	// Must follow the install-created regeneration rather than run alongside it:
	// a regeneration reads the latest state, fetches only its own partials and
	// saves a new row, so two in flight at once means the last writer drops the
	// other's partials from the current state.
	return regenerateState(ctx, regenerateStateRequest{
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintRunnerUpdated, ""),
		TriggeredByID:   install.RunnerID,
		TriggeredByType: "runners",
		MetricsWriter:   s.metrics,
	})
}

type regenerateStateRequest struct {
	InstallID       string
	Targets         []statemanager.PartialTarget
	TriggeredByID   string
	TriggeredByType string
	MetricsWriter   metrics.Writer
}

// regenerateState generates state in-band, always on the state-gen-v2 path. It
// replaces a state-partial-generate signal enqueued to the install's
// state-manager queue, which cost a queue hop each way to run the same code this
// workflow can run directly.
func regenerateState(ctx workflow.Context, req regenerateStateRequest) error {
	if err := statesignals.RegenerateWithMetrics(ctx, statesignals.RegenerateWithMetricsRequest{
		InstallID:       req.InstallID,
		Targets:         req.Targets,
		TriggeredByID:   req.TriggeredByID,
		TriggeredByType: req.TriggeredByType,
		MetricsWriter:   req.MetricsWriter,
	}); err != nil {
		return errors.Wrap(err, "unable to regenerate install state")
	}

	return nil
}

func enqueueRunnerServiceAccount(ctx workflow.Context, runnerID string) error {
	if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   runnerID,
		OwnerType: "runners",
		Signal: &runnersignals.Signal{
			RunnerID: runnerID,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to enqueue provision service account signal to runner")
	}

	return nil
}
