package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)


func (w *Workflows) deprovisionLegacy(ctx workflow.Context, org *app.Org, sandboxMode bool) error {
	_, err := w.execDeprovisionWorkflow(ctx, sandboxMode, &orgsv1.DeprovisionRequest{
		OrgId:  org.ID,
		Region: defaultOrgRegion,
	})
	if err != nil {
		w.updateStatus(ctx, org.ID, app.OrgStatusError, "unable to deprovision organization resources")
		return fmt.Errorf("unable to deprovision org: %w", err)
	}

	return nil
}
