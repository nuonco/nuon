package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

func (w *Workflows) provision(ctx workflow.Context, orgID string, sandboxMode bool) error {
	w.updateStatus(ctx, orgID, app.OrgStatusProvisioning, "provisioning organization resources")

	org, err := activities.AwaitGetByOrgID(ctx, orgID)
	if err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// NOTE(jm): this will be removed once the runner is in prod
	// and all orgs are migrated.
	if org.OrgType != app.OrgTypeV2 {
		if err := w.provisionLegacy(ctx, org, sandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org provision: %w", err)
		}

		return nil
	}

	// provision IAM roles for the org
	orgIAMReq := &executors.ProvisionIAMRequest{
		OrgId: orgID,
	}
	var orgIAMResp executors.ProvisionIAMResponse
	if err := w.execChildWorkflow(ctx, orgID, executors.ProvisionIAMWorkflowName, sandboxMode, orgIAMReq, &orgIAMResp); err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to reprovision iam roles")
		return fmt.Errorf("unable to provision iam roles: %w", err)
	}

	// provision the runner
	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationProvision,
	})
	// query runner until active

	w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
		OrgID:       orgID,
		SandboxMode: sandboxMode,
	})
	w.updateStatus(ctx, orgID, app.OrgStatusActive, "organization resources are provisioned")
	return nil
}
