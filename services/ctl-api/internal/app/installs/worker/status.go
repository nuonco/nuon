package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	erroractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/errors/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/structured_errors"
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

// clearDeployErrors clears any existing structured errors on a deploy, typically at the start of a retry.
func (w *Workflows) clearDeployErrors(ctx workflow.Context, deployID string) {
	l := workflow.GetLogger(ctx)
	if err := erroractivities.AwaitClearDeployErrors(ctx, erroractivities.ClearErrorsRequest{
		ID: deployID,
	}); err != nil {
		l.Error("unable to clear deploy errors",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}

// appendDeployErrors appends structured errors to a deploy (e.g. variable rendering failures).
func (w *Workflows) appendDeployErrors(ctx workflow.Context, deployID string, errs structured_errors.CompositeErrors) {
	if len(errs) == 0 {
		return
	}
	l := workflow.GetLogger(ctx)
	if err := erroractivities.AwaitAppendDeployErrors(ctx, erroractivities.AppendErrorsRequest{
		ID:     deployID,
		Errors: errs,
	}); err != nil {
		l.Error("unable to append deploy errors",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}

// clearBuildErrors clears any existing structured errors on a build.
func (w *Workflows) clearBuildErrors(ctx workflow.Context, buildID string) {
	l := workflow.GetLogger(ctx)
	if err := erroractivities.AwaitClearBuildErrors(ctx, erroractivities.ClearErrorsRequest{
		ID: buildID,
	}); err != nil {
		l.Error("unable to clear build errors",
			zap.String("build-id", buildID),
			zap.Error(err))
	}
}

// appendBuildErrors appends structured errors to a build.
func (w *Workflows) appendBuildErrors(ctx workflow.Context, buildID string, errs structured_errors.CompositeErrors) {
	if len(errs) == 0 {
		return
	}
	l := workflow.GetLogger(ctx)
	if err := erroractivities.AwaitAppendBuildErrors(ctx, erroractivities.AppendErrorsRequest{
		ID:     buildID,
		Errors: errs,
	}); err != nil {
		l.Error("unable to append build errors",
			zap.String("build-id", buildID),
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
