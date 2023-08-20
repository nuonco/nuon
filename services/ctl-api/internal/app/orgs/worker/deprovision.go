package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, orgID string, dryRun bool) error {
	// update status
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            "deprovisioning",
		StatusDescription: "tearing down org server, runner and infra - this should take about 15 minutes",
	}); err != nil {
		return fmt.Errorf("unable to update org status: %w", err)
	}

	_, err := w.execDeprovisionWorkflow(ctx, dryRun, &orgsv1.DeprovisionRequest{
		OrgId:  orgID,
		Region: defaultOrgRegion,
	})
	if err != nil {
		return fmt.Errorf("unable to deprovision org: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		OrgID: orgID,
	}); err != nil {
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}
