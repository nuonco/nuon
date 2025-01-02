package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	logv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/log/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) executeSandboxRun(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, op app.RunnerJobOperationType, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	// create the job
	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.Org.RunnerGroup.Runners[0].ID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   installRun.ID,
		Op:        op,
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	// check permissions
	var checkResp executors.CheckPermissionsResponse
	checkReq := &executors.CheckPermissionsRequest{
		AWSSettings:   w.protos.ToAWSSettings(install),
		AzureSettings: w.protos.ToAzureSettings(install),
		Metadata: executors.Metadata{
			OrgID:     install.OrgID,
			AppID:     install.AppID,
			InstallID: install.ID,
		},
	}

	if !sandboxMode {
		if err := w.execChildWorkflow(ctx, install.ID, executors.CheckPermissionsWorkflowName, sandboxMode, checkReq, &checkResp); err != nil {
			w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusAccessError, "unable to validate credentials before install "+err.Error())
			return fmt.Errorf("unable to validate credentials before install: %w", err)
		}
	} else {
		l.Info("skipping check permissions",
			zap.Any("install_id", install.ID),
			zap.String("org_id", install.OrgID))
	}

	// create the sandbox plan request
	planReq, err := w.protos.ToInstallPlanRequest(install, installRun.ID)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return fmt.Errorf("unable to get install plan request: %w", err)
	}
	planReq.LogConfiguration = &logv1.LogConfiguration{
		RunnerId: install.Org.RunnerGroup.Runners[0].ID,
		// RunnerApiToken: token.Token,
		RunnerApiUrl: w.cfg.RunnerAPIURL,
		RunnerJobId:  runnerJob.ID,
		Attrs:        logv1.NewAttrs(generics.ToStringMap(install.Org.RunnerGroup.Settings.Metadata)),
	}

	// create the sandbox plan
	planWorkflowID := fmt.Sprintf("%s-plan", installRun.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, planWorkflowID, planReq)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan")
		return fmt.Errorf("unable to create install plan: %w", err)
	}

	// store the plan in the db
	planJSON, err := protos.ToJSON(planResp.Plan)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to convert plan to json")
		return fmt.Errorf("unable to convert plan to json: %w", err)
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
	w.evClient.Send(ctx, install.Org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type:  runnersignals.OperationProcessJob,
		JobID: runnerJob.ID,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to poll job: %w", err)
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
