package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 20m
// @task-timeout 10m
func (w *Workflows) Reprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.OrgStatusProvisioning, "reprovisioning organization resources")

	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	// NOTE(jm): this will be removed once the runner is in prod and all orgs are migrated.
	if org.OrgType != app.OrgTypeV2 {
		if err := w.reprovisionLegacy(ctx, org, sreq.SandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org provision: %w", err)
		}

		return nil
	}

	// reprovision IAM roles for the org
	orgIAMReq := &executors.ProvisionIAMRequest{
		OrgID:       sreq.ID,
		Reprovision: true,
	}
	_, err = executors.AwaitProvisionIAM(ctx, orgIAMReq)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to reprovision iam roles")
		return fmt.Errorf("unable to reprovision iam roles: %w", err)
	}

	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationReprovision,
	})
	if err := w.pollRunner(ctx, org.RunnerGroup.Runners[0].ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "organization did not provision runner")
		return fmt.Errorf("runner did not reprovision correctly: %w", err)
	}

	w.updateStatus(ctx, sreq.ID, app.OrgStatusActive, "organization resources are provisioned")
	return nil
}
