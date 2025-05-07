package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) execCancelled(ctx workflow.Context, installWorkflowID string) error {
	wkflow, err := activities.AwaitGetInstallWorkflowByID(ctx, installWorkflowID)
	if err != nil {
		return err
	}

	switch wkflow.Type {
	case app.InstallWorkflowTypeReprovision:
		return w.cancelledInstallWorkflowTypeReprovision(ctx, wkflow)
	case app.InstallWorkflowTypeReprovisionSandbox:
		return w.cancelledInstallWorkflowTypeReprovisionSandbox(ctx, wkflow)
	}

	return nil
}

func (w *Workflows) cancelledInstallWorkflowTypeReprovisionSandbox(ctx workflow.Context, wkflow *app.InstallWorkflow) error {
	return w.updateSandboxRunStatusByInstallID(ctx, wkflow.InstallID)
}

func (w *Workflows) cancelledInstallWorkflowTypeReprovision(ctx workflow.Context, wkflow *app.InstallWorkflow) error {
	return w.updateSandboxRunStatusByInstallID(ctx, wkflow.InstallID)
}

func (w *Workflows) updateSandboxRunStatusByInstallID(ctx workflow.Context, installID string) error {
	installSandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return err
	}

	if installSandbox.Status != app.InstallSandboxStatusProvisioning && installSandbox.Status != app.InstallSandboxStatusDeprovisioning {
		return nil
	}
	return activities.AwaitUpdateRunStatusByInstallID(ctx, activities.UpdateRunStatusByInstallIDRequest{
		InstallID:         installID,
		Status:            app.SandboxRunStatusError,
		StatusDescription: "workflow cancelled",
	})
}
