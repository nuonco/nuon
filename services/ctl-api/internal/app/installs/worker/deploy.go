package worker

import (
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

func (w *Workflows) isBuildDeployable(bld *app.ComponentBuild) bool {
	return bld.Status == app.ComponentBuildStatusActive
}

func (w *Workflows) pollForDeployableBuild(ctx workflow.Context, installDeployId string, bld app.ComponentBuild) error {
	if w.isBuildDeployable(&bld) {
		return nil
	}
	sleepTimer := time.Second * 10
	maxAttempts := 20
	attempt := 0
	for {
		if attempt >= maxAttempts {
			return fmt.Errorf("build is not deployable after %d polling attempts", maxAttempts)
		}

		attempt++

		// Get the latest build
		bld, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, bld.ID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		// Check if the build is deployable
		if w.isBuildDeployable(bld) {
			return nil
		}

		if bld.Status == app.ComponentBuildStatusError {
			return fmt.Errorf("component build is in an error state")
		}

		workflow.Sleep(ctx, sleepTimer)
	}
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

		if app.InstallDeployStatus(depCmp.Component.Status) != app.InstallDeployStatusActive {
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
	w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusStarted)

	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install: %w", err)
	}
	defer func() {
		if pan := recover(); pan != nil {
			w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "internal error")
			w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusFailed)
			panic(pan)
		}
	}()

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: sreq.DeployID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("performing deploy")
	err = w.doDeploy(ctx, sreq, install)
	if err != nil {
		w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusFailed)
		w.sendNotification(ctx, notifications.NotificationsTypeDeployFailed, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})
	}
	return err
}

func (w *Workflows) doDeploy(ctx workflow.Context, sreq signals.RequestSignal, install *app.Install) error {
	installID := sreq.ID
	deployID := sreq.DeployID
	sandboxMode := sreq.SandboxMode

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, deployID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to get install deploy from database")
		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	err = w.pollForDeployableBuild(ctx, deployID, installDeploy.ComponentBuild)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusNoop, "build is not deployable")
		w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusNoop)
		return nil
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

	if err := w.execSync(ctx, install, installDeploy, sandboxMode); err != nil {
		return err
	}

	if err := w.execDeploy(ctx, install, installDeploy, sandboxMode); err != nil {
		return err
	}

	w.writeDeployEvent(ctx, deployID, signals.OperationDeploy, app.OperationStatusFinished)

	finalStatus := app.InstallDeployStatusActive
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		finalStatus = app.InstallDeployStatusInactive
	}
	w.updateDeployStatus(ctx, deployID, finalStatus, "deploy job finished")

	return nil
}
