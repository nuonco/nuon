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
)

func (w *Workflows) reprovision(ctx workflow.Context, installID string, dryRun bool) error {
	install, err := activities.AwaitGetByInstallID(ctx, installID)
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

	req, err := w.protos.ToInstallProvisionRequest(install, installRun.ID)
	if err != nil {
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
				w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusAccessError, "unable to assume provided role to access account")
				return fmt.Errorf("unable to reprovision install: %w", err)
			}
		}

		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to provision install resources")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to provision install: %w", err)
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "install resources provisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFinished)

	return nil
}
