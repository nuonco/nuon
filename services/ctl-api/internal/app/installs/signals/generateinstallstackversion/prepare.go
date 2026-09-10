package generateinstallstackversion

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	workerstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
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
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)

	// The service account is independent of state generation and of the stack
	// render below, so it overlaps them rather than serialising as its own step.
	saDone := workflow.NewChannel(ctx)
	var saErr error
	workflow.Go(ctx, func(gCtx workflow.Context) {
		saErr = provisionRunnerServiceAccount(gCtx, install, stateGenV2)
		saDone.Send(gCtx, true)
	})

	if err := generateInstallCreatedState(ctx, install.ID, stateGenV2, s.metrics); err != nil {
		saDone.Receive(ctx, nil)
		return err
	}

	saDone.Receive(ctx, nil)
	if saErr != nil {
		return saErr
	}

	return nil
}

func generateInstallCreatedState(ctx workflow.Context, installID string, stateGenV2 bool, mw metrics.Writer) error {
	if !stateGenV2 {
		if _, err := workerstate.AwaitGenerateState(ctx, &workerstate.GenerateStateRequest{
			InstallID:       installID,
			TriggeredByID:   installID,
			TriggeredByType: "installs",
		}); err != nil {
			return errors.Wrap(err, "unable to generate state")
		}
		return nil
	}

	if err := statesignals.RegenerateWithMetrics(ctx, statesignals.RegenerateWithMetricsRequest{
		InstallID:       installID,
		Targets:         statemanager.TargetsForHint(statemanager.HintInstallCreated, ""),
		TriggeredByID:   installID,
		TriggeredByType: "installs",
		MetricsWriter:   mw,
	}); err != nil {
		return errors.Wrap(err, "unable to regenerate install state")
	}

	return nil
}

func provisionRunnerServiceAccount(ctx workflow.Context, install *app.Install, stateGenV2 bool) error {
	if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   install.RunnerID,
		OwnerType: "runners",
		Signal: &runnersignals.Signal{
			RunnerID: install.RunnerID,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to enqueue provision service account signal to runner")
	}

	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		StateGenV2:      stateGenV2,
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintRunnerUpdated, ""),
		ForceAll:        true,
		TriggeredByID:   install.RunnerID,
		TriggeredByType: "runners",
	}); err != nil {
		return err
	}

	return nil
}
