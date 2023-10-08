package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reprovision(ctx workflow.Context, orgID string, dryRun bool) error {
	w.updateStatus(ctx, orgID, StatusProvisioning, "reprovisioning organization resources")

	_, err := w.execProvisionWorkflow(ctx, dryRun, &orgsv1.ProvisionRequest{
		OrgId:       orgID,
		Region:      defaultOrgRegion,
		Reprovision: true,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to reprovision organization resources")
		return fmt.Errorf("unable to provision org: %w", err)
	}

	w.updateStatus(ctx, orgID, StatusActive, "organization resources are provisioned")
	return nil
}
