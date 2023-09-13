package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, orgID string, dryRun bool) error {
	w.updateStatus(ctx, orgID, StatusDeprovisioning, "deprovisioning organization resources")

	_, err := w.execDeprovisionWorkflow(ctx, dryRun, &orgsv1.DeprovisionRequest{
		OrgId:  orgID,
		Region: defaultOrgRegion,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to deprovision organization resources")
		return fmt.Errorf("unable to deprovision org: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		OrgID: orgID,
	}); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to delete organiztion from database")
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}
