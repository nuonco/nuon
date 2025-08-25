package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/iam"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Provision(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)

	w.updateStatus(ctx, sreq.ID, app.OrgStatusProvisioning, "provisioning organization resources")

	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	// provision IAM roles for the org
	orgIAMReq := &orgiam.ProvisionIAMRequest{
		OrgID: sreq.ID,
	}
	if org.OrgType == app.OrgTypeDefault {
		_, err = orgiam.AwaitProvisionIAM(ctx, orgIAMReq)
		if err != nil {
			w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to provision IAM")
			return fmt.Errorf("unable to provision IAM: %w", err)
		}
	} else {
		l.Info("skipping await provision iam",
			zap.Any("org_type", org.OrgType),
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name))
	}

	// provision the runner
	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationProvision,
	})
	if err := w.pollRunner(ctx, org.RunnerGroup.Runners[0].ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "organization did not provision runner")
		return fmt.Errorf("runner did not provision correctly: %w", err)
	}

	w.updateStatus(ctx, sreq.ID, app.OrgStatusActive, "organization resources are provisioned")
	return nil
}
