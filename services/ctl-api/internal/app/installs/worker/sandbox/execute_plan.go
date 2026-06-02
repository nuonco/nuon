package sandbox

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/workflowstepapprovalrequest"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) executeSandboxPlan(ctx workflow.Context, install *app.Install, sandboxRun *app.InstallSandboxRun, stepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	op := app.RunnerJobOperationTypeCreateApplyPlan
	if sandboxRun.RunType == app.SandboxRunTypeDeprovision {
		op = app.RunnerJobOperationTypeCreateTeardownPlan
	}

	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.RunnerID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   sandboxRun.ID,
		JobType:   install.AppSandboxConfig.JobType(),
		Op:        op,
		Metadata: map[string]string{
			"install_id":       install.ID,
			"sandbox_run_id":   sandboxRun.ID,
			"sandbox_run_type": string(sandboxRun.RunType),
		},
	})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, sandboxRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	planResponse, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      sandboxRun.ID,
		InstallID:  install.ID,
		RootDomain: w.cfg.DNSRootDomain,
		WorkflowID: fmt.Sprintf("%s-create-api-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, sandboxRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	compositePlan := plantypes.CompositePlan{
		SandboxRunPlan: planResponse.Plan,
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, sandboxRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     install.ID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, sandboxRun.ID, app.SandboxRunStatusError, "unable to record install role usage")
		return fmt.Errorf("unable to record install role usage: %w", err)
	}

	// queue job
	l.Info("queued job and waiting on it to be picked up by runner")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   install.RunnerID,
		WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateRunStatusWithoutStatusSync(ctx, sandboxRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to execute job: %w", err)
	}
	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		w.updateRunStatusWithoutStatusSync(ctx, sandboxRun.ID, app.SandboxRunStatusError, "job failed with status"+string(status))
	}

	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job")
	}

	var installWorkflowID string
	if sandboxRun.WorkflowID != nil {
		installWorkflowID = *sandboxRun.WorkflowID
	}
	if err := workflowstepapprovalrequest.Dispatch(ctx, &workflowstepapprovalrequest.Signal{
		InstallID:         install.ID,
		InstallWorkflowID: installWorkflowID,
		WorkflowStepID:    stepID,
		OwnerID:           sandboxRun.ID,
		OwnerType:         "install_sandbox_runs",
		RunnerJobID:       job.ID,
		ApprovalType:      app.TerraformPlanApprovalType,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	return nil
}
