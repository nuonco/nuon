package builds

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/queuebuild"
	componentdeploysyncandplan "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeploysyncandplan"
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

	// Load the branch to get install groups for sandbox build
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	l.Info("triggering builds",
		"app_branch_id", s.AppBranchID,
		"app_config_id", run.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

	if len(appConfig.ComponentIDs) == 0 {
		l.Info("no components to build")
		return nil
	}

	if len(branch.Configs) == 0 {
		return fmt.Errorf("app branch %s has no configs", s.AppBranchID)
	}

	config := branch.Configs[0]

	// Launch buildComponents and sandboxBuild in parallel
	errCh := workflow.NewChannel(ctx)

	workflow.Go(ctx, func(gCtx workflow.Context) {
		buildErr := s.buildComponents(gCtx, l, appConfig.ComponentIDs, run.AppConfigID)
		errCh.Send(gCtx, buildErr)
	})

	workflow.Go(ctx, func(gCtx workflow.Context) {
		sandboxErr := s.sandboxBuild(gCtx, l, appConfig.ComponentIDs, &config)
		errCh.Send(gCtx, sandboxErr)
	})

	// Collect 2 results
	var errs []error
	for i := 0; i < 2; i++ {
		var resultErr error
		errCh.Receive(ctx, &resultErr)
		if resultErr != nil {
			errs = append(errs, resultErr)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("builds had %d errors: %v", len(errs), errs)
	}

	l.Info("all builds completed successfully")
	return nil
}

// buildComponents enqueues build signals for each component in parallel and waits for completion.
func (s *Signal) buildComponents(ctx workflow.Context, l log.Logger, componentIDs []string, appConfigID string) error {
	errCh := workflow.NewChannel(ctx)
	pending := len(componentIDs)

	for _, componentID := range componentIDs {
		componentID := componentID
		workflow.Go(ctx, func(gCtx workflow.Context) {
			buildErr := s.buildComponent(gCtx, l, componentID, appConfigID)
			errCh.Send(gCtx, buildErr)
		})
	}

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

func (s *Signal) buildComponent(ctx workflow.Context, l log.Logger, componentID, appConfigID string) error {
	// Enqueue a queue_build signal to the component's queue
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   componentID,
		OwnerType: "components",
		Signal: &queuebuild.Signal{
			ComponentID: componentID,
			AppConfigID: appConfigID,
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

// sandboxBuild triggers sandbox deploy-sync-and-plan for each install component across all install groups.
func (s *Signal) sandboxBuild(ctx workflow.Context, l log.Logger, componentIDs []string, config *app.AppBranchConfig) error {
	if len(config.InstallGroups) == 0 {
		l.Info("no install groups for sandbox build, skipping")
		return nil
	}

	for _, group := range config.InstallGroups {
		if len(group.InstallIDs) == 0 {
			l.Info("no installs in group, skipping", "group_name", group.Name)
			continue
		}

		// Deploy to each install in the group in parallel
		errCh := workflow.NewChannel(ctx)
		pending := len(group.InstallIDs)

		for _, installID := range group.InstallIDs {
			installID := installID
			workflow.Go(ctx, func(gCtx workflow.Context) {
				deployErr := s.sandboxBuildForInstall(gCtx, l, installID, componentIDs)
				errCh.Send(gCtx, deployErr)
			})
		}

		var errs []error
		for i := 0; i < pending; i++ {
			var deployErr error
			errCh.Receive(ctx, &deployErr)
			if deployErr != nil {
				errs = append(errs, deployErr)
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("sandbox build group %s had %d errors: %v", group.Name, len(errs), errs)
		}
	}

	l.Info("sandbox builds completed successfully")
	return nil
}

// sandboxBuildForInstall triggers sandbox componentdeploysyncandplan for each component in an install.
func (s *Signal) sandboxBuildForInstall(ctx workflow.Context, l log.Logger, installID string, componentIDs []string) error {
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

	for _, componentID := range componentIDs {
		installComponentID, ok := installComponentMap[componentID]
		if !ok {
			l.Warn("install component not found for sandbox build, skipping",
				"install_id", installID,
				"component_id", componentID,
			)
			continue
		}

		enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   installID,
			OwnerType: "installs",
			Signal: &componentdeploysyncandplan.Signal{
				SandboxMode:        true,
				ComponentID:        componentID,
				InstallComponentID: installComponentID,
			},
		})
		if err != nil {
			return fmt.Errorf("install %s component %s: enqueue failed: %w", installID, componentID, err)
		}

		l.Info("waiting for sandbox build to complete",
			"install_id", installID,
			"component_id", componentID,
			"install_component_id", installComponentID,
			"queue_signal_id", enqueueResp.QueueSignalID,
		)

		_, err = queueclient.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
		if err != nil {
			return fmt.Errorf("install %s component %s: sandbox build failed: %w", installID, componentID, err)
		}

		l.Info("sandbox build completed",
			"install_id", installID,
			"component_id", componentID,
		)
	}

	return nil
}
