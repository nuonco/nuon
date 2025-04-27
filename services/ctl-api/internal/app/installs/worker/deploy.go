package worker

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

func (w *Workflows) isDeployable(install *app.Install) bool {
	return install.InstallSandboxRuns[0].Status == app.SandboxRunStatusActive
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
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}
	defer func() {
		if pan := recover(); pan != nil {
			w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "internal error")
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
		w.sendNotification(ctx, notifications.NotificationsTypeDeployFailed, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})
	}
	return err
}

func (w *Workflows) doDeploy(ctx workflow.Context, sreq signals.RequestSignal, install *app.Install) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	installID := sreq.ID
	deployID := sreq.DeployID
	sandboxMode := sreq.SandboxMode

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, deployID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to get install deploy from database")
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	err = w.pollForDeployableBuild(ctx, deployID, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusNoop, "build is not deployable")
		return nil
	}

	if installDeploy.Type == app.InstallDeployTypeTeardown && !sreq.ForceDelete {
		if !w.isTeardownable(install) {
			l.Error("component is not in a status to be torn down")
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "install is not in a delete_queued, deprovisioning or active state to tear down components")
			return nil
		}

		// check if the component is a dependency of another component that is still active
		invertedDepIds, err := activities.AwaitFetchUntornDependencies(ctx, activities.FetchUntornDependenciesRequest{
			ComponentRootID: installDeploy.ComponentID,
			InstallID:       installID,
		})
		if err != nil {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to check dependencies")
			return fmt.Errorf("unable to fetch active inverted dependencies: %w", err)
		}

		if len(invertedDepIds) > 0 {
			l.Error("component is depended on by other components")
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, fmt.Sprintf("compoent is depended on by orher components IDs: [%s]", strings.Join(invertedDepIds, ", ")))
			return fmt.Errorf("other components depends on this component depIDs: %s", strings.Join(invertedDepIds, ", "))
		}
	}

	if installDeploy.Type != app.InstallDeployTypeTeardown {
		if !w.isDeployable(install) {
			l.Error("install is not currently deployable, due to its status")
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "install is not active and can not be deployed too")
			return nil
		}

		inactiveDepIDs, err := activities.AwaitFetchInactiveDependencies(ctx, activities.FetchInactiveDependenciesRequest{
			ComponentRootID: installDeploy.ComponentID,
			InstallID:       installID,
		})
		if err != nil {
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, "unable to check dependencies")
			return fmt.Errorf("unable to check dependencies: %w", err)
		}

		if len(inactiveDepIDs) > 0 {
			// TODO(jm): ask robisso why we aren't using the stuff in `deploy_dependencies.go`
			l.Error("dependent component was not active: " + inactiveDepIDs[0])
			w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, fmt.Sprintf("dependent component: [%s]  not active", strings.Join(inactiveDepIDs, ", ")))
			return fmt.Errorf("dependent component: [%s]  not active", strings.Join(inactiveDepIDs, ", "))
		}
	}

	// skip lifecycle hooks if the deploy is a teardown
	if installDeploy.Type != app.InstallDeployTypeTeardown {
		if err := w.AwaitLifecycleActionWorkflows(ctx, &LifecycleActionWorkflowsRequest{
			InstallID:       install.ID,
			TriggerType:     app.ActionWorkflowTriggerTypePreDeployAll,
			TriggeredByID:   installDeploy.ID,
			TriggeredByType: "install_deploys",
			RunEnvVars: generics.ToPtrStringMap(map[string]string{
				"TRIGGER":        string(app.ActionWorkflowTriggerTypePreDeployAll),
				"DEPLOY_TYPE":    string(installDeploy.Type),
				"DEPLOY_ID":      installDeploy.ID,
				"COMPONENT_ID":   installDeploy.InstallComponent.ID,
				"COMPONENT_NAME": installDeploy.InstallComponent.Component.Name,
			}),
		}); err != nil {
			return errors.Wrap(err, "lifecycle pre hooks failed")
		}
	}

	if err := w.execSync(ctx, install, installDeploy, sandboxMode); err != nil {
		return w.errorResponse(ctx, sreq, deployID, installDeploy.InstallComponentID, "error syncing", err)
	}

	if err := w.execDeploy(ctx, install, installDeploy, sandboxMode); err != nil {
		return w.errorResponse(ctx, sreq, deployID, installDeploy.InstallComponentID, "error deploying", err)
	}

	// skip lifecycle hooks if the deploy is a teardown
	if installDeploy.Type != app.InstallDeployTypeTeardown {
		// run hooks after the deploy
		if err := w.AwaitLifecycleActionWorkflows(ctx, &LifecycleActionWorkflowsRequest{
			InstallID:       install.ID,
			TriggerType:     app.ActionWorkflowTriggerTypePostDeployAll,
			TriggeredByID:   installDeploy.ID,
			TriggeredByType: "install_deploys",
			RunEnvVars: generics.ToPtrStringMap(map[string]string{
				"TRIGGER":        string(app.ActionWorkflowTriggerTypePostDeployAll),
				"DEPLOY_TYPE":    string(installDeploy.Type),
				"DEPLOY_ID":      installDeploy.ID,
				"COMPONENT_ID":   installDeploy.InstallComponent.ID,
				"COMPONENT_NAME": installDeploy.InstallComponent.Component.Name,
			}),
		}); err != nil {
			return errors.Wrap(err, "lifecycle post hooks failed")
		}
	}

	finalDeployStatus := app.InstallDeployStatusActive
	finalComponentStatus := app.InstallComponentStatusActive
	finalMessage := "deploy job finished"
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		finalDeployStatus = app.InstallDeployStatusInactive
		finalComponentStatus = app.InstallComponentStatusDeleted
	}
	if sreq.ForceDelete {
		w.updateDeployStatusWithoutStatusSync(ctx, deployID, finalDeployStatus, finalMessage)
		w.updateInstallComponentStatus(ctx, installDeploy.InstallComponentID, finalComponentStatus, finalMessage)
		if err := activities.AwaitDeleteInstallComponent(ctx, activities.DeleteInstallComponentRequest{
			InstallComponentID: installDeploy.InstallComponentID,
		}); err != nil {
			return errors.Wrap(err, "unable to delete install component")
		}

	} else {
		w.updateDeployStatus(ctx, deployID, finalDeployStatus, finalMessage)
	}

	return nil
}

func (w *Workflows) errorResponse(ctx workflow.Context, sreq signals.RequestSignal, deployID, installComponentID, message string, err error) error {
	if sreq.ForceDelete {
		w.updateDeployStatusWithoutStatusSync(ctx, deployID, app.InstallDeployStatusInactive, message)
		w.updateInstallComponentStatus(ctx, installComponentID, app.InstallComponentStatusDeleteFailed, message)
		if err := activities.AwaitDeleteInstallComponent(ctx, activities.DeleteInstallComponentRequest{
			InstallComponentID: installComponentID,
		}); err != nil {
			return errors.Wrap(err, "unable to delete install component")
		}
		return nil
	}

	w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, message)
	return errors.Wrap(err, message)
}
