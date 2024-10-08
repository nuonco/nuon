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
)

func (w *Workflows) provisionLegacy(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, sandboxMode bool) error {
	req, err := w.protos.ToInstallProvisionRequest(install, installRun.ID)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install provision request")
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	_, err = w.execProvisionWorkflow(ctx, sandboxMode, req)
	if err != nil {
		w.mw.Event(ctx, &statsd.Event{
			Title: "install failed to provision",
			Text: fmt.Sprintf(
				"install %s failed to provision\ncreated by %s\nerror: %s",
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
			if strings.Contains(err.Error(), accessError.Error()) || strings.Contains(err.Error(), "iam-role") {
				w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusAccessError, "unable to assume provided role to access account")
				return fmt.Errorf("unable to provision install: %w", err)
			}
		}

		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to provision install resources")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationProvision, app.OperationStatusFailed)
		return fmt.Errorf("unable to provision install: %w", err)
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "install resources provisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationProvision, app.OperationStatusFinished)

	return nil
}
