package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

func (w *Workflows) reprovisionLegacy(ctx workflow.Context, org *app.Org, sandboxMode bool) error {
	_, err := w.execDeprovisionIAMWorkflow(ctx, sandboxMode, &executors.DeprovisionIAMRequest{
		OrgID: org.ID,
	})

	// NOTE(jm): we ignore errors deprovisioning, and make a best effort to reprovision regardless
	_, err = w.execProvisionWorkflow(ctx, sandboxMode, &orgsv1.ProvisionRequest{
		OrgId:       org.ID,
		Region:      defaultOrgRegion,
		Reprovision: true,
		CustomCert:  false,
	})
	if err != nil {
		w.updateStatus(ctx, org.ID, app.OrgStatusError, "unable to reprovision organization resources")
		return fmt.Errorf("unable to reprovision org: %w", err)
	}

	for _, orgApp := range org.Apps {
		if err := activities.AwaitUpsertProject(ctx, activities.UpsertProjectRequest{
			OrgID:     org.ID,
			ProjectID: orgApp.ID,
		}); err != nil {
			w.updateStatus(ctx, org.ID, app.OrgStatusError, "unable to upsert app project")
			return fmt.Errorf("unable to reprovision org: %w", err)
		}

		for _, install := range orgApp.Installs {
			if err := activities.AwaitUpsertProject(ctx, activities.UpsertProjectRequest{
				OrgID:     org.ID,
				ProjectID: install.ID,
			}); err != nil {
				w.updateStatus(ctx, org.ID, app.OrgStatusError, "unable to upsert install project")
				return fmt.Errorf("unable to reprovision org: %w", err)
			}
		}
	}

	return nil
}
