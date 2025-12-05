package actions

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (w *Workflows) updateStatus(ctx workflow.Context, runID string, status app.InstallActionWorkflowRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateInstallWorkflowRunStatus(ctx, activities.UpdateInstallWorkflowRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status",
			zap.String("run-id", runID),
			zap.Error(err))
	}
	if err := statusactivities.AwaitUpdateInstallWorkflowRunStatusV2(ctx, statusactivities.UpdateInstallWorkflowRunStatusV2Request{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status v2",
			zap.String("run-id", runID),
			zap.Error(err))
	}
}

// TODO(sdboyer) refactor this to return an error; processing should abort if status updates fail
func (w *Workflows) updateRunStatus(ctx workflow.Context, runID string, status app.SandboxRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateRunStatus(ctx, activities.UpdateRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update run status",
			zap.String("run-id", runID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateRunStatusV2(ctx, statusactivities.UpdateRunStatusV2Request{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status v2",
			zap.String("run-id", runID),
			zap.Error(err))
	}
}

func (w *Workflows) updateRunStatusWithoutStatusSync(ctx workflow.Context, runID string, status app.SandboxRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateRunStatus(ctx, activities.UpdateRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update run status",
			zap.String("run-id", runID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateRunStatusV2(ctx, statusactivities.UpdateRunStatusV2Request{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update run status v2",
			zap.String("run-id", runID),
			zap.Error(err))
	}

}

func (w *Workflows) updateInstallSandboxStatus(ctx workflow.Context, runID string, status app.InstallSandboxStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateInstallSandboxStatus(ctx, activities.UpdateInstallSandboxStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update install sandbox status",
			zap.String("run-id", runID),
			zap.Error(err))
	}
}

func (w *Workflows) updateDeployStatus(ctx workflow.Context, deployID string, status app.InstallDeployStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: statusDescription,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status v2",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}

func (w *Workflows) updateDeployStatusWithoutStatusSync(ctx workflow.Context, deployID string, status app.InstallDeployStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: statusDescription,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status v2",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}

func (w *Workflows) updateInstallComponentStatus(ctx workflow.Context, installComponentID string, status app.InstallComponentStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateInstallComponentStatus(ctx, activities.UpdateInstallComponentStatusRequest{
		InstallComponentID: installComponentID,
		Status:             status,
		StatusDescription:  statusDescription,
	}); err != nil {
		l.Error("unable to update indtall component status",
			zap.String("InstallComponentID", installComponentID),
			zap.Error(err))
	}
}

func (w *Workflows) updateActionRunStatus(ctx workflow.Context, runID string, status app.InstallActionWorkflowRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateInstallWorkflowRunStatus(ctx, activities.UpdateInstallWorkflowRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status",
			zap.String("run-id", runID),
			zap.Error(err))
	}
	if err := statusactivities.AwaitUpdateInstallWorkflowRunStatusV2(ctx, statusactivities.UpdateInstallWorkflowRunStatusV2Request{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status v2",
			zap.String("run-id", runID),
			zap.Error(err))
	}
}
