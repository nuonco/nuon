package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) ensureOrg(ctx workflow.Context, appID string) error {
	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.GetOrg, activities.GetOrgRequest{
		AppID: appID,
	}, &org); err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	if org.Status != "active" {
		return fmt.Errorf("org is not active: %s", org.Status)
	}

	return nil
}
