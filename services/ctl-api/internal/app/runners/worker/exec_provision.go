package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	kuberunner "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/kuberunner"
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

	req := &kuberunner.ProvisionRunnerRequest{
		RunnerID:                 runnerID,
		APIURL:                   runner.RunnerGroup.Settings.RunnerAPIURL,
		APIToken:                 apiToken,
		RunnerIAMRole:            fmt.Sprintf("arn:aws:iam::%s:role/orgs/%s/runner-%s", w.cfg.ManagementAccountID, runnerID, runnerID),
		RunnerServiceAccountName: runner.RunnerGroup.Settings.OrgK8sServiceAccountName,
		Image: kuberunner.ProvisionRunnerRequestImage{
			URL: runner.RunnerGroup.Settings.ContainerImageURL,
			Tag: runner.RunnerGroup.Settings.ContainerImageTag,
		},
	}
	_, err = kuberunner.AwaitProvisionRunner(ctx, req)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to provision runner")
		return errors.Wrap(err, "unable to provision runner")
	}

	w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "runner is active and ready to process jobs")
	return nil
}
