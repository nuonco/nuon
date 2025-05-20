package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
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
		RunnerIAMRole:            runner.RunnerGroup.Settings.OrgAWSIAMRoleARN,
		RunnerServiceAccountName: runner.RunnerGroup.Settings.OrgK8sServiceAccountName,
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
