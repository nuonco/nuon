package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 30s
func (w *Workflows) Seed(ctx workflow.Context, _ signals.RequestSignal) error {
	if w.cfg.Env == "prod" || w.cfg.Env == "production" {
		return errors.New("seeding is not supported in prod")
	}

	orgs, err := activities.AwaitGetOrgs(ctx, activities.GetOrgsRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to get orgs")
	}

	for _, org := range orgs {
		w.seedOrg(ctx, org.ID)
	}

	return nil
}

func (w *Workflows) seedOrg(ctx workflow.Context, orgID string) error {
	// restart the org, force it into sandbox mode and reprovsion it
	w.ev.Send(ctx, orgID, &orgssignals.Signal{
		Type: orgssignals.OperationRestart,
	})
	w.ev.Send(ctx, orgID, &orgssignals.Signal{
		Type: orgssignals.OperationForceSandboxMode,
	})
	w.ev.Send(ctx, orgID, &orgssignals.Signal{
		Type: orgssignals.OperationStageSeed,
	})
	w.ev.Send(ctx, orgID, &orgssignals.Signal{
		Type: orgssignals.OperationRestart,
	})
	return nil
}
