package deploygrouptoqueue

import (
	"fmt"
	"log"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	// Load the run to get the AppConfigID
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	// Get the install group
	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return fmt.Errorf("unable to get install group: %w", err)
	}

	// Resolve install IDs: use label selector if set, otherwise static IDs
	installIDs := group.InstallIDs
	if group.LabelSelector != nil {
		branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
		if err != nil {
			return fmt.Errorf("unable to get app branch for label resolution: %w", err)
		}

		resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
			AppID:    branch.AppID,
			GroupID:  group.ID,
			Selector: group.LabelSelector,
		})
		if err != nil {
			return fmt.Errorf("unable to resolve install group labels: %w", err)
		}
		installIDs = resolved.InstallIDs

		logger.Info("resolved install group via label selector",
			"install_group_id", group.ID,
			"resolved_count", len(installIDs),
		)
	}

	logger.Info("deploying install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(installIDs),
		"max_parallel", group.MaxParallel,
	)

	if len(installIDs) == 0 {
		logger.Info("no installs in group, skipping")
		return nil
	}

	// Create an install config update workflow per install and wait for all to complete
	errCh := workflow.NewChannel(ctx)
	pending := len(installIDs)

	for _, installID := range installIDs {
		installID := installID
		workflow.Go(ctx, func(gCtx workflow.Context) {
			result, createErr := activities.AwaitCreateInstallConfigUpdateWorkflow(gCtx, &activities.CreateInstallConfigUpdateWorkflowInput{
				InstallID:      installID,
				NewAppConfigID: run.AppConfigID,
				AppBranchRunID: s.RunID,
				InstallGroupID: group.ID,
				PlanOnly:       run.PlanOnly,
			})
			if createErr != nil {
				errCh.Send(gCtx, fmt.Errorf("install %s: unable to create config update workflow: %w", installID, createErr))
				return
			}

			logger.Info("created install config update workflow",
				"install_id", installID,
				"workflow_id", result.WorkflowID,
				"install_config_update_id", result.InstallConfigUpdateID,
			)

			// The install workflow is now managed by the conductor - it will execute
			// the generated steps (diff, deploy changed components, etc.)
			// We don't need to wait for completion here since the conductor handles it.
			errCh.Send(gCtx, nil)
		})
	}

	// Collect results
	var errs []error
	for i := 0; i < pending; i++ {
		var deployErr error
		errCh.Receive(ctx, &deployErr)
		if deployErr != nil {
			errs = append(errs, deployErr)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("deploy group had %d errors: %v", len(errs), errs)
	}

	logger.Info("install group deployment initiated successfully",
		"install_group_id", group.ID,
	)

	return nil
}

// deployToInstall deploys all components to a single install via cross-namespace queue signals.
// For each component, it looks up the install component, enqueues a componentdeploysyncandplan signal
// to the install's queue, and waits for completion.
func (s *Signal) deployToInstall(ctx workflow.Context, logger log.Logger, installID string, componentIDs []string) error {
	// Look up install component mappings for this install
	mappingResp, err := activities.AwaitGetInstallComponentsByComponentIDs(ctx, activities.GetInstallComponentsByComponentIDsRequest{
		Req: &activities.GetInstallComponentsByComponentIDsInput{
			InstallID:    installID,
			ComponentIDs: componentIDs,
		},
	})
	if err != nil {
		return fmt.Errorf("install %s: unable to get install components: %w", installID, err)
	}

	// Build a map of componentID -> installComponentID for quick lookup
	installComponentMap := make(map[string]string, len(mappingResp.Mappings))
	for _, m := range mappingResp.Mappings {
		installComponentMap[m.ComponentID] = m.InstallComponentID
	}

	// Deploy each component to this install sequentially
	// (within an install, components should be deployed in order)
	for _, componentID := range componentIDs {
		installComponentID, ok := installComponentMap[componentID]
		if !ok {
			logger.Warn("install component not found, skipping",
				"install_id", installID,
				"component_id", componentID,
			)
			continue
		}

		err := s.deployComponentToInstall(ctx, logger, installID, componentID, installComponentID)
		if err != nil {
			return fmt.Errorf("install %s component %s: %w", installID, componentID, err)
		}
	}

	logger.Info("all components deployed to install", "install_id", installID)
	return nil
}

// deployComponentToInstall enqueues a componentdeploysyncandplan signal to the install's queue
// and waits for it to complete.
func (s *Signal) deployComponentToInstall(ctx workflow.Context, logger log.Logger, installID, componentID, installComponentID string) error {
	cb := callback.New(ctx, installComponentID)
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   installID,
		OwnerType: "installs",
		QueueName: "install-signals",
		Signal: &componentdeploysyncandplan.Signal{
			InstallComponentID: installComponentID,
			ComponentID:        componentID,
		},
		Callback: cb,
	})
	if err != nil {
		return fmt.Errorf("enqueue failed: %w", err)
	}

	logger.Info("waiting for component deploy to complete",
		"install_id", installID,
		"component_id", componentID,
		"install_component_id", installComponentID,
		"queue_signal_id", enqueueResp.QueueSignalID,
	)

	_, err = callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout)
	if err != nil {
		return fmt.Errorf("deploy failed: %w", err)
	}

	logger.Info("component deploy completed",
		"install_id", installID,
		"component_id", componentID,
	)

	return nil
}
