package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reprovision(ctx workflow.Context, orgID string, dryRun bool) error {
	w.updateStatus(ctx, orgID, StatusProvisioning, "reprovisioning organization resources")

	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: orgID,
	}, &org); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	_, err := w.execDeprovisionIAMWorkflow(ctx, dryRun, &iamv1.DeprovisionIAMRequest{
		OrgId: orgID,
	})
	// NOTE(jm): we ignore errors deprovisioning, and make a best effort to reprovision regardless

	_, err = w.execProvisionWorkflow(ctx, dryRun, &orgsv1.ProvisionRequest{
		OrgId:       orgID,
		Region:      defaultOrgRegion,
		Reprovision: true,
		CustomCert:  org.CustomCert,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to reprovision organization resources")
		return fmt.Errorf("unable to provision org: %w", err)
	}

	if err := w.defaultExecGetActivity(ctx, w.acts.ReprovisionApps, activities.ReprovisionAppsRequest{
		OrgID: orgID,
	}, &org); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to reprovision apps")
		return fmt.Errorf("unable to reprovision apps: %w", err)
	}

	w.updateStatus(ctx, orgID, StatusActive, "organization resources are provisioned")
	return nil
}
