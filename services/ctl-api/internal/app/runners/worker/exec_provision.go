package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	logv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/log/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) executeProvisionOrgRunner(ctx workflow.Context, runnerID, apiToken string, sandboxMode bool) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "local runner must be run locally")
		return nil
	}
	if runner.Org.OrgType == app.OrgTypeIntegration {
		w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "integration mode, bypassing provisioning")
		return nil
	}

	req := &executors.ProvisionRunnerRequest{
		RunnerID:                 runnerID,
		APIURL:                   runner.RunnerGroup.Settings.RunnerAPIURL,
		APIToken:                 apiToken,
		RunnerIAMRole:            runner.RunnerGroup.Settings.AWSIAMRoleARN,
		RunnerServiceAccountName: runner.RunnerGroup.Settings.K8sServiceAccountName,
		Image: executors.ProvisionRunnerRequestImage{
			URL: runner.RunnerGroup.Settings.ContainerImageURL,
			Tag: runner.RunnerGroup.Settings.ContainerImageTag,
		},
	}
	var resp executors.ProvisionRunnerResponse
	err = w.execChildWorkflow(ctx, runnerID, executors.ProvisionRunnerWorkflowName, sandboxMode, req, &resp)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to provision runner")
		return fmt.Errorf("unable to provision runner: %w", err)
	}

	w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "runner is active and ready to process jobs")
	return nil
}

func (w *Workflows) executeProvisionInstallRunner(ctx workflow.Context, runnerID, apiToken string, sandboxMode bool) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "local install")
		return nil
	}
	if runner.RunnerGroup.Platform == app.AppRunnerTypeAWS {
		w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "aws install")
		return nil
	}

	if runner.Org.OrgType == app.OrgTypeIntegration {
		return nil
	}

	install, err := activities.AwaitGetInstall(ctx, activities.GetInstallRequest{
		InstallID: runner.RunnerGroup.OwnerID,
	})
	if err != nil {
		return fmt.Errorf("unable to get runner install: %w", err)
	}

	token, err := activities.AwaitCreateToken(ctx, activities.CreateTokenRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create token")
	}

	logStream, err := cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		return errors.Wrap(err, "no log stream found")
	}

	// create the job
	runnerJob, err := activities.AwaitCreateJob(ctx, &activities.CreateJobRequest{
		RunnerID:    runner.Org.RunnerGroup.Runners[0].ID,
		OwnerType:   "runners",
		OwnerID:     runnerID,
		Op:          app.RunnerJobOperationTypeCreate,
		Type:        runner.RunnerGroup.Platform.JobType(),
		LogStreamID: logStream.ID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create provision job")
		return fmt.Errorf("unable to create job: %w", err)
	}

	// create the sandbox plan request
	planWorkflowID := fmt.Sprintf("%s-runner-provision", runnerID)
	planReq, err := w.protos.ToRunnerInstallPlanRequest(runner, install, apiToken)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create runner plan request")
		return fmt.Errorf("unable to create runner plan: %w", err)
	}
	planReq.LogConfiguration = &logv1.LogConfiguration{
		RunnerId:       install.RunnerGroup.Runners[0].ID,
		RunnerApiToken: token.Token,
		RunnerApiUrl:   w.cfg.RunnerAPIURL,
		// RunnerJobId:    logStreamID,
		Attrs: logv1.NewAttrs(generics.ToStringMap(runner.RunnerGroup.Settings.Metadata)),
	}

	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, planWorkflowID, planReq)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create provision plan")
		return fmt.Errorf("unable to create runner plan: %w", err)
	}

	// store the plan in the db
	planJSON, err := protos.ToJSON(planResp.Plan)
	if err != nil {
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create plan json")
		return fmt.Errorf("unable to save runner job plan: %w", err)
	}

	// queue job
	w.evClient.Send(ctx, runner.Org.RunnerGroup.Runners[0].ID, &signals.Signal{
		Type:  signals.OperationProcessJob,
		JobID: runnerJob.ID,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to poll runner job to completion")
		return fmt.Errorf("unable to poll runner job to completion: %w", err)
	}

	w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "runner is active and ready to process jobs")
	return nil
}
