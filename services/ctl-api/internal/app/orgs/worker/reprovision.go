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
// @execution-timeout 20m
// @task-timeout 10m
func (w *Workflows) Reprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)
	w.updateStatus(ctx, sreq.ID, app.OrgStatusProvisioning, "reprovisioning organization resources")

	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	// deprovision IAM roles for the org
	if org.OrgType == app.OrgTypeDefault {
		orgIAMReq := &orgiam.DeprovisionIAMRequest{
			OrgID: sreq.ID,
		}

		_, err = orgiam.AwaitDeprovisionIAM(ctx, orgIAMReq)
		if err != nil {
			w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to deprovision iam roles")
			return fmt.Errorf("unable to deprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await deprovision iam",
			zap.Any("org_type", org.OrgType),
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name))
	}

	// provision IAM roles for the org
	if org.OrgType == app.OrgTypeDefault {
		orgIAMReq := &orgiam.ProvisionIAMRequest{
			OrgID:       sreq.ID,
			Reprovision: true,
		}
		_, err = orgiam.AwaitProvisionIAM(ctx, orgIAMReq)
		if err != nil {
			w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to reprovision iam roles")
			return fmt.Errorf("unable to reprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await reprovision iam",
			zap.Any("org_type", org.OrgType),
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name))
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
