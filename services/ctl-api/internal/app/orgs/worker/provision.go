package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, orgID string, dryRun bool) error {
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            "provisioning",
		StatusDescription: "creating org management server, build runner and infra - this should take about 15 minutes",
	}); err != nil {
		return fmt.Errorf("unable to update org status: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, dryRun, &orgsv1.ProvisionRequest{
		OrgId:       orgID,
		Region:      defaultOrgRegion,
		Reprovision: false,
	})
	if err != nil {
		return fmt.Errorf("unable to provision org: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            "active",
		StatusDescription: "active",
	}); err != nil {
		return fmt.Errorf("unable to update org status: %w", err)
	}
	return nil
}
