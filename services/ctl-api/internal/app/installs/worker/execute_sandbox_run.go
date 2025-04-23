package worker

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) executeSandboxRun(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, op app.RunnerJobOperationType, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureIndependentRunner))
	if err != nil {
		return err
	}

	// create the job
	targetRunnerID := install.Org.RunnerGroup.Runners[0].ID
	if enabled {
		targetRunnerID = install.RunnerID
	}

	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  targetRunnerID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   installRun.ID,
		Op:        op,
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

	enabled, err = activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureIndependentRunner))
	if err != nil {
		return errors.Wrap(err, "unable to check runner feature")
	}

	runPlan, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      installRun.ID,
		InstallID:  install.ID,
		WorkflowID: fmt.Sprintf("%s-create-api-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return errors.Wrap(err, "unable to create plan")
	}

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
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   targetRunnerID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to execute job: %w", err)
	}

	l.Info("configuring DNS for nuon.run domain if enabled")
	dnsReq := &executors.ProvisionDNSDelegationRequest{
		Metadata: &executors.Metadata{
			OrgID:     install.OrgID,
			AppID:     install.AppID,
			InstallID: install.ID,
		},
	}

	if op == app.RunnerJobOperationTypeDestroy {
		return nil
	}

	l.Info("provisioning nuon.run root domain")
	if !sandboxMode {
		_, err = executors.AwaitProvisionDNSDelegation(ctx, dnsReq)
		if err != nil {
			w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to provision dns delegation")
			return errors.Wrap(err, "unable to provision dns delegation")
		}
		l.Info("successfully provisioned dns delegation")
	} else {
		l.Info("skipping dns delegation provisioning",
			zap.Any("install_id", install.ID),
			zap.String("org_id", install.OrgID))
	}

	return nil
}
