package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

func (w *Workflows) delete(ctx workflow.Context, orgID string, dryRun bool) error {
	if err := w.deprovision(ctx, orgID, dryRun); err != nil {
		return err
	}

	// update status with response
	if err := activities.AwaitDeleteByOrgID(ctx, orgID); err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to delete organization from database")
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}
