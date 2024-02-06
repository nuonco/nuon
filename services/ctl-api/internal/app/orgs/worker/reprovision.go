package worker

import (
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reprovision(ctx workflow.Context, orgID string, sandboxMode bool) error {
	w.updateStatus(ctx, orgID, StatusProvisioning, "reprovisioning organization resources")

	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: orgID,
	}, &org); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	_, err := w.execDeprovisionIAMWorkflow(ctx, sandboxMode, &iamv1.DeprovisionIAMRequest{
		OrgId: orgID,
	})
	// NOTE(jm): we ignore errors deprovisioning, and make a best effort to reprovision regardless

	_, err = w.execProvisionWorkflow(ctx, sandboxMode, &orgsv1.ProvisionRequest{
		OrgId:       orgID,
		Region:      defaultOrgRegion,
		Reprovision: true,
		CustomCert:  org.CustomCert,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to reprovision organization resources")
		return fmt.Errorf("unable to reprovision org: %w", err)
	}

	for _, app := range org.Apps {
		if err := w.defaultExecGetActivity(ctx, w.acts.UpsertProject, activities.UpsertProjectRequest{
			OrgID:     orgID,
			ProjectID: app.ID,
		}, &org); err != nil {
			w.updateStatus(ctx, orgID, StatusError, "unable to upsert app project")
			return fmt.Errorf("unable to reprovision org: %w", err)
		}

		for _, install := range app.Installs {
			if err := w.defaultExecGetActivity(ctx, w.acts.UpsertProject, activities.UpsertProjectRequest{
				OrgID:     orgID,
				ProjectID: install.ID,
			}, &org); err != nil {
				w.updateStatus(ctx, orgID, StatusError, "unable to upsert install project")
				return fmt.Errorf("unable to reprovision org: %w", err)
			}
		}
	}

	w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
		OrgID:       orgID,
		SandboxMode: sandboxMode,
	})

	w.updateStatus(ctx, orgID, StatusActive, "organization resources are provisioned")
	return nil
}
