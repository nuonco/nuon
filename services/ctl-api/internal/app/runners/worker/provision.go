package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

func (w *Workflows) provision(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	w.updateStatus(ctx, runnerID, app.RunnerStatusProvisioning, "provisioning organization resources")

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	_, err = activities.AwaitCreateAccount(ctx, activities.CreateAccountRequest{
		RunnerID: runnerID,
		OrgID:    runner.ID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create runner service account")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	token, err := activities.AwaitCreateToken(ctx, activities.CreateTokenRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create runner token")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		return w.executeProvisionOrgRunner(ctx, runnerID, token.Token, sandboxMode)
	case app.RunnerGroupTypeInstall:
		return w.executeProvisionInstallRunner(ctx, runnerID, token.Token, sandboxMode)
	}

	return nil
}
