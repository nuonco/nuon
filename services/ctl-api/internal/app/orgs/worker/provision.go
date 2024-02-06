package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, orgID string, sandboxMode bool) error {
	w.updateStatus(ctx, orgID, StatusProvisioning, "provisioning organization resources")

	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: orgID,
	}, &org); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to get org from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, sandboxMode, &orgsv1.ProvisionRequest{
		OrgId:       orgID,
		Region:      defaultOrgRegion,
		Reprovision: false,
		CustomCert:  org.CustomCert,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to provision organization resources")
		return fmt.Errorf("unable to provision org: %w", err)
	}

	w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
		OrgID:       orgID,
		SandboxMode: sandboxMode,
	})

	w.updateStatus(ctx, orgID, StatusActive, "organization resources are provisioned")
	return nil
}
