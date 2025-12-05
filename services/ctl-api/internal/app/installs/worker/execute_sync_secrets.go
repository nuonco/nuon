package worker

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @execution-timeout 30m
func (w *Workflows) SyncSecrets(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: sreq.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		StepID: sreq.WorkflowStepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if sreq.SandboxMode {
		l.Debug("skipping sync secrets in sandbox mode")
		return nil
	}

	l.Info("creating plan")
	plan, err := plan.AwaitCreateSyncSecretsPlan(ctx, &plan.CreateSyncSecretsPlanRequest{
		InstallID:  install.ID,
		WorkflowID: fmt.Sprintf("%s-create-sync-secrets-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create plan")
	}

	if len(plan.KubernetesSecrets) < 1 {
		l.Debug("no secrets to sync")
		return nil
	}

	// create the job
	l.Info("creating sync secrets job")
	runnerJob, err := activities.AwaitCreateSyncSecretsJob(ctx, &activities.CreateSyncSecretsJobRequest{
		RunnerID:  install.RunnerID,
		InstallID: install.ID,
		OwnerType: "install_workflow_steps",
		OwnerID:   sreq.WorkflowStepID,
		Op:        app.RunnerJobOperationTypeExec,
		Metadata: map[string]string{
			"install_id": install.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	// Deprecated: for now we dual write both the plan json and the composite plan
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncSecretsPlan: plan,
		},
	}); err != nil {
		return fmt.Errorf("unable to save plan: %w", err)
	}

	l.Info("queueing job and waiting on execution")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   install.RunnerID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		return fmt.Errorf("unable to sync secrets: %w", err)
	}

	return nil
}
