package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) isDeprovisionable(ctx workflow.Context, install *app.Install) (bool, []zap.Field, error) {
	attributes := []zap.Field{
		zap.String("sandbox.run_count", fmt.Sprintf("%d", len(install.InstallSandboxRuns))),
		zap.String("sandbox.status", fmt.Sprintf("%s", install.InstallSandboxRuns[0].Status)),
	}

	// NOTE(fd): we may want to not fetch all of them in w/e query sets this val
	if len(install.InstallSandboxRuns) < 1 {
		attributes = append(attributes, zap.String("reason", fmt.Sprintf("fewer than 1 sandbox run")))
		return false, attributes, nil
	}

	if install.InstallSandboxRuns[0].Status == app.SandboxRunStatusAccessError {
		attributes = append(attributes, zap.String("reason", fmt.Sprintf("sandbox status is in error")))
		return false, attributes, nil
	}

	untornInstallDeploys, err := activities.AwaitFetchUntornInstallDeploys(ctx, activities.FetchUntornInstallDeploysRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return false, attributes, fmt.Errorf("unable to fetch untorn install deploys: %w", err)
	}

	if len(untornInstallDeploys) > 0 {
		untornInlComponentIDs := []string{}
		untornAppComponentIDs := []string{}
		for _, idpl := range untornInstallDeploys {
			untornInlComponentIDs = append(untornInlComponentIDs, idpl.InstallComponentID)
			untornAppComponentIDs = append(untornAppComponentIDs, idpl.ComponentID)

		}
		attributes = append(attributes, zap.Strings("install.untorn_component_ids", untornInlComponentIDs))
		attributes = append(attributes, zap.Strings("install.untorn_app_component_ids", untornAppComponentIDs))
		attributes = append(attributes, zap.String("reason", fmt.Sprintf("at least one install component cannot be torn down")))
		return false, attributes, nil
	}

	return true, attributes, nil
}

// @temporal-gen workflow
// @execution-timeout 60m
// @execution-timeout 30m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	sandboxMode := sreq.SandboxMode

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeDeprovision,
	})
	if err != nil {
		return fmt.Errorf("unable to create install: %w", err)
	}
	enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureIndependentRunner))
	if err != nil {
		return err
	}
	if enabled {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         sreq.WorkflowStepID,
			StepTargetID:   installRun.ID,
			StepTargetType: plugins.TableName(w.db, installRun),
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	defer func() {
		if pan := recover(); pan != nil {
			w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "internal error")
			w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
			panic(pan)
		}
	}()

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		SandboxRunID: installRun.ID,
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
	l.Info("deprovisioning install")

	if !enabled {
		isDeprovisionable, attributes, err := w.isDeprovisionable(ctx, install)
		if err != nil {
			l.Error("unable to determine if install is deprovisionable", attributes...)

			w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to determine if install is deprovisionable")
			w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
			return fmt.Errorf("unable to determine if install is deprovisionable: %w", err)
		}

		if !isDeprovisionable {
			l.Error("install is not deprovisionable, this will be a NOOP", attributes...)
			w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "install is not deprovisionable")
			w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusNoop)
			return nil
		}
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusDeprovisioning, "deprovisioning")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusStarted)

	// deprovision the runner
	if !enabled {
		l.Info("starting runner deprovision")
		w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
			Type: runnersignals.OperationDeprovision,
		})

		// wait until the runner is deprovisioned
		if err := w.pollRunnerDeprovisioned(ctx, install.RunnerGroup.Runners[0].ID); err != nil {
			l.Error("runner was unable to be deprovisioned correctly. Continuing to deprovision sandbox", zap.Error(err))
		}
	}

	// wait for the runner
	l.Info("executing deprovision")
	err = w.executeSandboxRun(ctx, install, installRun, app.RunnerJobOperationTypeDestroy, sandboxMode)
	if err != nil {
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
		return err
	}

	l.Info("deprovision was successful")
	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusDeprovisioned, "successfully deprovisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFinished)
	return nil
}
