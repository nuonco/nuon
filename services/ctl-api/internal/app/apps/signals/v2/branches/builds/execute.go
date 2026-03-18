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

	l.Info("triggering builds",
		"app_branch_id", s.AppBranchID,
		"app_config_id", run.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

	if len(appConfig.ComponentIDs) == 0 {
		l.Info("no components to build")
		return nil
	}

	// Step 1: Build all components (must complete before sandbox deploys can use the artifacts)
	if err := s.buildComponents(ctx, l, appConfig.ComponentIDs, run.AppConfigID); err != nil {
		return fmt.Errorf("component builds failed: %w", err)
	}

	// Step 2: Sandbox build — deploy components to sandbox installs using built artifacts
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if len(branch.Configs) == 0 {
		l.Info("no branch config found, skipping sandbox build")
		return nil
	}

	if err := s.sandboxBuild(ctx, l, appConfig.ComponentIDs, &branch.Configs[0]); err != nil {
		return fmt.Errorf("sandbox builds failed: %w", err)
	}

	l.Info("all builds completed successfully")
	return nil
}

// buildComponents enqueues build signals for each component in parallel and waits for all to complete.
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
		return fmt.Errorf("%d component build(s) failed: %v", len(errs), errs)
	}

	l.Info("all component builds completed successfully")
	return nil
}

func (s *Signal) buildComponent(ctx workflow.Context, l log.Logger, componentID, appConfigID string) error {
	if err := activities.AwaitEnsureComponentQueueByComponentID(ctx, componentID); err != nil {
		return fmt.Errorf("component %s: ensure queue failed: %w", componentID, err)
	}

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

	if _, err = queueclient.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID); err != nil {
		return fmt.Errorf("component %s: build failed: %w", componentID, err)
	}

	l.Info("component build completed", "component_id", componentID)
	return nil
}

// sandboxBuild deploys built components to sandbox installs across all install groups in parallel.
func (s *Signal) sandboxBuild(ctx workflow.Context, l log.Logger, componentIDs []string, config *app.AppBranchConfig) error {
	if len(config.InstallGroups) == 0 {
		l.Info("no install groups, skipping sandbox build")
		return nil
	}

	for _, group := range config.InstallGroups {
		if len(group.InstallIDs) == 0 {
			l.Info("no installs in group, skipping", "group_name", group.Name)
			continue
		}

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
			return fmt.Errorf("sandbox build group %q had %d error(s): %v", group.Name, len(errs), errs)
		}
	}

	l.Info("sandbox builds completed successfully")
	return nil
}

// sandboxBuildForInstall triggers componentdeploysyncandplan (sandbox mode) for each component in an install.
func (s *Signal) sandboxBuildForInstall(ctx workflow.Context, l log.Logger, installID string, componentIDs []string) error {
	mappingResp, err := activities.AwaitGetInstallComponentsByComponentIDs(ctx, activities.GetInstallComponentsByComponentIDsRequest{
		Req: &activities.GetInstallComponentsByComponentIDsInput{
			InstallID:    installID,
			ComponentIDs: componentIDs,
		},
	})
	if err != nil {
		return fmt.Errorf("install %s: unable to get install components: %w", installID, err)
	}

	installComponentMap := make(map[string]string, len(mappingResp.Mappings))
	for _, m := range mappingResp.Mappings {
		installComponentMap[m.ComponentID] = m.InstallComponentID
	}

	for _, componentID := range componentIDs {
		installComponentID, ok := installComponentMap[componentID]
		if !ok {
			l.Warn("install component mapping not found, skipping",
				"install_id", installID,
				"component_id", componentID)
			continue
		}

		enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   installID,
			OwnerType: "installs",
			QueueName: "install-signals",
			Signal: &componentdeploysyncandplan.Signal{
				SandboxMode:        true,
				ComponentID:        componentID,
				InstallComponentID: installComponentID,
			},
		})
		if err != nil {
			return fmt.Errorf("install %s component %s: enqueue failed: %w", installID, componentID, err)
		}

		l.Info("waiting for sandbox deploy to complete",
			"install_id", installID,
			"component_id", componentID,
			"install_component_id", installComponentID,
			"queue_signal_id", enqueueResp.QueueSignalID)

		if _, err = queueclient.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID); err != nil {
			return fmt.Errorf("install %s component %s: sandbox deploy failed: %w", installID, componentID, err)
		}

		l.Info("sandbox deploy completed", "install_id", installID, "component_id", componentID)
	}

	return nil
}
