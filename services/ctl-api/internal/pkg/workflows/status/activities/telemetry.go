package statusactivities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (a *Activities) logStepStatus(ctx context.Context, step app.WorkflowStep, previous, status app.CompositeStatus) {
	if previous.Status == status.Status {
		return
	}

	var flowEvent string
	switch status.Status {
	case app.StatusError:
		flowEvent = "step.errored"
	case app.StatusSuccess:
		flowEvent = "step.completed"
	default:
		return
	}

	fields := []zap.Field{
		zap.String("flow_event", flowEvent),
		zap.String("org_id", step.OrgID),
		zap.String("workflow_id", step.InstallWorkflowID),
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.String("owner_id", step.OwnerID),
		zap.String("owner_type", step.OwnerType),
		zap.Int("step_idx", step.Idx),
		zap.Int("group_idx", step.GroupIdx),
		zap.Int("group_retry_idx", step.GroupRetryIdx),
		zap.Int("retry_index", step.RetryIndex),
		zap.String("status", string(status.Status)),
		zap.Any("status_metadata", status.Metadata),
	}
	if step.OwnerType == "installs" {
		fields = append(fields, zap.String("install_id", step.OwnerID))
	}
	if status.StatusHumanDescription != "" {
		fields = append(fields, zap.String("status_description", status.StatusHumanDescription))
	}

	l := cctx.GetLogger(ctx, a.l)
	if status.Status == app.StatusError {
		l.Error("flow telemetry", append(fields, zap.String("error", status.StatusHumanDescription))...)
		return
	}
	l.Info("flow telemetry", fields...)
}

func (a *Activities) logWorkflowError(ctx context.Context, wf app.Workflow, previous, status app.CompositeStatus) {
	if status.Status != app.StatusError || previous.Status == status.Status {
		return
	}
	fields := []zap.Field{
		zap.String("flow_event", "workflow.failed"),
		zap.String("org_id", wf.OrgID),
		zap.String("workflow_id", wf.ID),
		zap.String("workflow_type", string(wf.Type)),
		zap.String("owner_id", wf.OwnerID),
		zap.String("owner_type", wf.OwnerType),
		zap.String("status", string(status.Status)),
		zap.String("error", status.StatusHumanDescription),
	}
	if wf.OwnerType == "installs" {
		fields = append(fields, zap.String("install_id", wf.OwnerID))
	}
	cctx.GetLogger(ctx, a.l).Error("flow telemetry", fields...)
}

func (a *Activities) logRunnerJob(ctx context.Context, job app.RunnerJob, status app.RunnerJobStatus, description string) {
	fields := []zap.Field{
		zap.String("flow_event", "runner_job."+string(status)),
		zap.String("runner_job_id", job.ID),
		zap.String("workflow_id", job.FlowWorkflowID()),
		zap.String("install_id", job.FlowInstallID()),
		zap.String("owner_id", job.OwnerID),
		zap.String("owner_type", job.OwnerType),
		zap.String("runner_id", job.RunnerID),
		zap.String("job_type", string(job.Type)),
		zap.String("job_operation", string(job.Operation)),
		zap.String("status", string(status)),
	}
	failed := false
	switch status {
	case app.RunnerJobStatusFailed, app.RunnerJobStatusTimedOut:
		failed = true
		fields = append(fields, zap.String("error", description))
	default:
		if description != "" {
			fields = append(fields, zap.String("status_description", description))
		}
	}

	l := cctx.GetLogger(ctx, a.l)
	if failed {
		l.Error("flow telemetry", fields...)
		return
	}
	l.Info("flow telemetry", fields...)
}
