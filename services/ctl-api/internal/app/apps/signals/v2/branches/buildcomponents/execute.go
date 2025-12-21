package buildcomponents

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/queuebuild"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Load the run to get the AppConfigID (set by the appconfig step)
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	// Get app config with component IDs
	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, run.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	l.Info("triggering component builds via queue signals",
		"app_branch_id", s.AppBranchID,
		"app_config_id", run.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

	if len(appConfig.ComponentIDs) == 0 {
		l.Info("no components to build")
		return nil
	}

	// Enqueue build signals to each component's queue in parallel using workflow.Go
	// Use a Temporal channel to collect results from each goroutine
	errCh := workflow.NewChannel(ctx)
	pending := len(appConfig.ComponentIDs)

	for _, componentID := range appConfig.ComponentIDs {
		componentID := componentID
		workflow.Go(ctx, func(gCtx workflow.Context) {
			buildErr := s.buildComponent(gCtx, l, componentID)
			errCh.Send(gCtx, buildErr)
		})
	}

	// Collect results from all goroutines
	var errs []error
	for i := 0; i < pending; i++ {
		var buildErr error
		errCh.Receive(ctx, &buildErr)
		if buildErr != nil {
			errs = append(errs, buildErr)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("component builds had %d errors: %v", len(errs), errs)
	}

	l.Info("all component builds completed successfully")
	return nil
}

func (s *Signal) buildComponent(ctx workflow.Context, l log.Logger, componentID string) error {
	// Enqueue a queue_build signal to the component's queue
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   componentID,
		OwnerType: "components",
		Signal: &queuebuild.Signal{
			ComponentID: componentID,
		},
	})
	if err != nil {
		return fmt.Errorf("component %s: enqueue failed: %w", componentID, err)
	}

	l.Info("waiting for component build to complete",
		"component_id", componentID,
		"queue_signal_id", enqueueResp.QueueSignalID)

	// Wait for the build signal to complete
	_, err = queueclient.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
	if err != nil {
		return fmt.Errorf("component %s: build failed: %w", componentID, err)
	}

	l.Info("component build completed", "component_id", componentID)
	return nil
}
