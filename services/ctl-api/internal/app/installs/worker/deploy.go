package worker

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) isBuildDeployable(bld app.ComponentBuild) bool {
	return bld.Status == app.ComponentBuildStatusActive
}

func (w *Workflows) isDeployable(install *app.Install) bool {
	return install.InstallSandboxRuns[0].Status == app.SandboxRunStatusActive
}

func (w *Workflows) anyDependencyInActive(ctx workflow.Context, install app.Install, installDeploy app.InstallDeploy) (string, error) {
	depComponents, err := activities.AwaitGetComponentDependents(ctx, activities.GetComponentDependents{
		AppID:           install.App.ID,
		ComponentRootID: installDeploy.ComponentID,
	})
	if err != nil {
		return "", fmt.Errorf("unable to get installComponent: %w", err)
	}

	for _, dep := range depComponents {
		var depCmp *app.InstallComponent
		depCmp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installDeploy.InstallComponent.InstallID,
			ComponentID: dep.ID,
		})

		if depCmp == nil {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("unable to get installComponent: %w", err)
		}

		if app.InstallDeployStatus(depCmp.Component.Status) != app.InstallDeployStatusOK {
			return depCmp.ComponentID, fmt.Errorf("dependent component: %s, not active", depCmp.ID)
		}
	}

	return "", nil
}

func (w *Workflows) isTeardownable(install *app.Install) bool {
	if install.InstallSandboxRuns[0].Status == app.SandboxRunStatusError {
		return false
	}

	if install.InstallSandboxRuns[0].Status == app.SandboxRunStatusAccessError {
		return false
	}

	return true
}

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Deploy(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	deployID := sreq.DeployID
	sandboxMode := sreq.SandboxMode

	w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusStarted)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to get install from database")
		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install: %w", err)
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, deployID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to get install deploy from database")
		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	org, err := activities.AwaitGetOrg(ctx, activities.GetOrgRequest{
		InstallID: installID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to get org from database")
		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get org: %w", err)
	}

	if installDeploy.Type == app.InstallDeployTypeTeardown {
		if !w.isTeardownable(install) {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "install is not in a delete_queued, deprovisioning or active state to tear down components")
			w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusNoop)
			return nil
		}

		// check if the component is a dependency of another component that is still active
		invertedDepIds, err := activities.AwaitFetchUntornDependencies(ctx, activities.FetchUntornDependenciesRequest{
			ComponentRootID: installDeploy.ComponentID,
			InstallID:       installID,
		})

		if err != nil {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to check dependencies")
			w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
			return fmt.Errorf("unable to fetch active inverted dependencies: %w", err)
		}

		if len(invertedDepIds) > 0 {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, fmt.Sprintf("compoent is depended on by orher components IDs: [%s]", strings.Join(invertedDepIds, ", ")))
			return fmt.Errorf("other components depends on this component depIDs: %s", strings.Join(invertedDepIds, ", "))
		}
	} else {
		if !w.isDeployable(install) {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "install is not active and can not be deployed too")
			w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusNoop)
			return nil
		}

		inactiveDepIDs, err := activities.AwaitFetchInactiveDependencies(ctx, activities.FetchInactiveDependenciesRequest{
			ComponentRootID: installDeploy.ComponentID,
			InstallID:       installID,
		})
		if err != nil {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to check dependencies")
			w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
			return fmt.Errorf("unable to check dependencies: %w", err)
		}

		if len(inactiveDepIDs) > 0 {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, fmt.Sprintf("dependent component: [%s]  not active", strings.Join(inactiveDepIDs, ", ")))
			w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
			return fmt.Errorf("dependent component: [%s]  not active", strings.Join(inactiveDepIDs, ", "))
		}
	}

	if !w.isBuildDeployable(installDeploy.ComponentBuild) {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusNoop, "build is not deployable")
		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
		return nil
	}

	if org.OrgType != app.OrgTypeV2 {
		if err := w.execSyncLegacy(ctx, install, installDeploy, sandboxMode); err != nil {
			return err
		}
		if err := w.execDeployLegacy(ctx, install, installDeploy, sandboxMode); err != nil {
			return err
		}

		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFinished)
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusOK, "deploy is active")
		return nil
	}

	if err := w.execSync(ctx, install, installDeploy, sandboxMode); err != nil {
		return err
	}
	if err := w.execDeploy(ctx, install, installDeploy, sandboxMode); err != nil {
		return err
	}

	w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFinished)
	w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusOK, "deploy is active")

	return nil
}
