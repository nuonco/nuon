package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Provision(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.RunnerStatusProvisioning, "provisioning organization resources")

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	_, err = activities.AwaitCreateAccount(ctx, activities.CreateAccountRequest{
		RunnerID: sreq.ID,
		OrgID:    runner.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create runner service account")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	token, err := activities.AwaitCreateToken(ctx, activities.CreateTokenRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create runner token")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		return w.executeProvisionOrgRunner(ctx, sreq.ID, token.Token, sreq.SandboxMode)
	case app.RunnerGroupTypeInstall:
		return w.executeProvisionInstallRunner(ctx, sreq.ID, token.Token, sreq.SandboxMode, sreq.LogStreamID)
	}

	return nil
}
