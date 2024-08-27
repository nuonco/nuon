package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) ensureOrg(ctx workflow.Context, appID string) error {
	if org, err := activities.AwaitGetOrgByAppID(ctx, appID); err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	} else if org.Status != "active" {
		return fmt.Errorf("org is not active: %s", org.Status)
	}

	return nil
}
