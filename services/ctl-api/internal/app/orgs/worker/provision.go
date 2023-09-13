package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, orgID string, dryRun bool) error {
	w.updateStatus(ctx, orgID, StatusProvisioning, "provisioning organization resources")

	_, err := w.execProvisionWorkflow(ctx, dryRun, &orgsv1.ProvisionRequest{
		OrgId:       orgID,
		Region:      defaultOrgRegion,
		Reprovision: false,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to provision organization resources")
		return fmt.Errorf("unable to provision org: %w", err)
	}

	w.updateStatus(ctx, orgID, StatusActive, "organization resources are provisioned")
	return nil
}
