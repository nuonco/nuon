package sandbox

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) executeApplyPlan(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, stepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	l.Info("executing apply plan")

	prevJob, err := activities.AwaitGetLatestJobByOwnerID(ctx, installRun.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get latest runner job")
	}

	// create the job
	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.RunnerID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   installRun.ID,
		Op:        app.RunnerJobOperationTypeApplyPlan,
		Metadata: map[string]string{
			"install_id":       install.ID,
			"sandbox_run_id":   installRun.ID,
			"sandbox_run_type": string(installRun.RunType),
		},
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	l.Info("creating sandbox run plan")
	runPlan, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      installRun.ID,
		InstallID:  install.ID,
		RootDomain: w.cfg.DNSRootDomain,
		WorkflowID: fmt.Sprintf("%s-create-api-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return errors.Wrap(err, "unable to create plan")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         stepID,
		StepTargetID:   installRun.ID,
		StepTargetType: plugins.TableName(w.db, installRun),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}
	runPlan.ApplyPlanContents = prevJob.Execution.Result.Contents
	runPlan.ApplyPlanDisplay = prevJob.Execution.Result.ContentsDisplay

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	l.Info("queued job and waiting on it to be picked up by runner event loop")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   install.RunnerID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to execute job: %w", err)
	}
	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job failed with status"+string(status))
	}

	// job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	// if err != nil {
	// 	return errors.Wrap(err, "unable to get job")
	// }

	// NOTE(fd): this is not necessary here - this is only necessary in the plan step
	// if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
	// 	OwnerID:     installRun.ID,
	// 	OwnerType:   "install_sandbox_runs",
	// 	RunnerJobID: job.ID,
	// 	StepID:      stepID,
	// 	Plan:        job.Execution.Result.Contents,
	// 	Type:        app.TerraformPlanApprovalType,
	// }); err != nil {
	// 	return errors.Wrap(err, "unable to create approval")
	// }

	return nil
}
