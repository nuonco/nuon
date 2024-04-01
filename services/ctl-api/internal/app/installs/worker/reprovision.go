package worker

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reprovision(ctx workflow.Context, installID string, dryRun bool) error {
	w.updateStatus(ctx, installID, StatusProvisioning, "reprovisioning install resources")

	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installRun app.InstallSandboxRun
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateSandboxRun, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeReprovision,
	}, &installRun); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to create sandbox run")
		w.updateRunStatus(ctx, installRun.ID, StatusError, "unable to create sandbox run")
		return fmt.Errorf("unable to create install: %w", err)
	}

	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusStarted)
	w.updateRunStatus(ctx, installRun.ID, StatusProvisioning, "provisioning")

	req, err := w.protos.ToInstallProvisionRequest(&install, installRun.ID)
	if err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to create install provision request")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	_, err = w.execProvisionWorkflow(ctx, dryRun, req)
	if err != nil {
		w.mw.Event(ctx, &statsd.Event{
			Title: "install failed to reprovision",
			Text: fmt.Sprintf(
				"install %s failed to reprovision\ncreated by %s\nerror: %s",
				installID,
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
				w.updateStatus(ctx, installID, StatusAccessError, "unable to assume provided role to access account")
				return fmt.Errorf("unable to reprovision install: %w", err)
			}
		}

		w.updateStatus(ctx, installID, StatusError, "unable to reprovision app resources")
		w.updateRunStatus(ctx, installRun.ID, StatusError, "unable to provision install resources")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to provision install: %w", err)
	}

	w.updateStatus(ctx, installID, StatusActive, "app resources are provisioned")
	w.updateRunStatus(ctx, installRun.ID, StatusActive, "install resources provisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFinished)

	return nil
}
