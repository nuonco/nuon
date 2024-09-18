package worker

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

func (w *Workflows) reprovisionLegacy(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, sandboxMode bool) error {
	req, err := w.protos.ToInstallProvisionRequest(install, installRun.ID)
	if err != nil {
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	_, err = w.execProvisionWorkflow(ctx, sandboxMode, req)
	if err != nil {
		w.mw.Event(ctx, &statsd.Event{
			Title: "install failed to reprovision",
			Text: fmt.Sprintf(
				"install %s failed to reprovision\ncreated by %s\nerror: %s",
				install.ID,
				install.CreatedBy.Email,
				err.Error(),
			),
			Tags: metrics.ToTags(map[string]string{
				"status":             "error",
				"status_description": "failed to provision",
			}),
		})

		if install.AWSAccount != nil {
			accessError := credentials.ErrUnableToAssumeRole{
				RoleARN: install.AWSAccount.IAMRoleARN,
			}
			if strings.Contains(err.Error(), accessError.Error()) {
				w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusAccessError, "unable to assume provided role to access account")
				return fmt.Errorf("unable to reprovision install: %w", err)
			}
		}

		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to provision install resources")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to provision install: %w", err)
	}

	return nil
}

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Reprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	sandboxMode := sreq.SandboxMode

	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeReprovision,
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create sandbox run")
		return fmt.Errorf("unable to create install: %w", err)
	}

	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusStarted)
	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusProvisioning, "provisioning")

	if install.Org.OrgType != app.OrgTypeV2 {
		if err := w.reprovisionLegacy(ctx, install, installRun, sandboxMode); err != nil {
			return err
		}

		return nil
	}

	if err := w.executeSandboxRun(ctx, install, installRun, app.RunnerJobOperationTypeCreate, sandboxMode); err != nil {
		return err
	}

	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationReprovision,
	})
	if err := w.pollRunner(ctx, install.RunnerGroup.Runners[0].ID); err != nil {
		return err
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "install resources provisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFinished)

	return nil
}
