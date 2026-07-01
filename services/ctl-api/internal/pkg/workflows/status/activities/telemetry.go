package statusactivities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// logStepError is the authoritative step-error surface: a step can fail and park
// (awaiting retry) without its signal returning, so the lifecycle hook never sees
// a terminal outcome — the error only exists on the step's status.
func (a *Activities) logStepError(ctx context.Context, step app.WorkflowStep, status app.CompositeStatus) {
	if status.Status != app.StatusError {
		return
	}
	cctx.GetLogger(ctx, a.l).Error("flow telemetry",
		zap.String("flow_event", "step.errored"),
		zap.String("workflow_id", step.InstallWorkflowID),
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.String("status", string(status.Status)),
		zap.String("error", status.StatusHumanDescription),
	)
}

func (a *Activities) logWorkflowError(ctx context.Context, wf app.Workflow, status app.CompositeStatus) {
	if status.Status != app.StatusError {
		return
	}
	fields := []zap.Field{
		zap.String("flow_event", "workflow.failed"),
		zap.String("workflow_id", wf.ID),
		zap.String("workflow_type", string(wf.Type)),
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
		zap.String("step_id", job.FlowStepID()),
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
