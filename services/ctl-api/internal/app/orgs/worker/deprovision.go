package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollAppsDeprovisioned(ctx workflow.Context, orgID string) error {
	for {
		org, err := activities.AwaitGetByOrgID(ctx, orgID)
		if err != nil {
			w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
			return fmt.Errorf("unable to get org: %w", err)
		}

		if len(org.Apps) < 1 {
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}
}

func (w *Workflows) deprovision(ctx workflow.Context, orgID string, sandboxMode bool) error {
	w.updateStatus(ctx, orgID, app.OrgStatusActive, "ensuring all apps are deleted before deprovisioning")
	if err := w.pollAppsDeprovisioned(ctx, orgID); err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "error polling apps being deprovisioned")
		return fmt.Errorf("unable to poll for deleted apps: %w", err)
	}

	return w.deprovisionOrg(ctx, orgID, sandboxMode)
}

func (w *Workflows) deprovisionOrg(ctx workflow.Context, orgID string, sandboxMode bool) error {
	org, err := activities.AwaitGet(ctx, activities.GetRequest{
		OrgID: orgID,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	w.updateStatus(ctx, orgID, app.OrgStatusDeprovisioning, "deprovisioning organization resources")

	// NOTE(jm): this will be removed once the runner is in prod and all orgs are migrated.
	if org.OrgType != app.OrgTypeV2 {
		if err := w.deprovisionLegacy(ctx, org, sandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org deprovision: %w", err)
		}

		return nil
	}

	// reprovision IAM roles for the org
	orgIAMReq := &executors.DeprovisionIAMRequest{
		OrgId: orgID,
	}
	var orgIAMResp executors.ProvisionIAMResponse
	if err := w.execChildWorkflow(ctx, orgID, executors.DeprovisionIAMWorkflowName, sandboxMode, orgIAMReq, &orgIAMResp); err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to deprovision iam roles")
		return fmt.Errorf("unable to deprovision iam roles: %w", err)
	}

	// TODO(jm): wait until this is deprovisioned
	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDeprovision,
	})
	return nil
}
