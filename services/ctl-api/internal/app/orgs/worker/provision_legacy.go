package worker

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/metrics"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (w *Workflows) provisionLegacy(ctx workflow.Context, org *app.Org, sandboxMode bool) error {
	_, err := w.execProvisionWorkflow(ctx, sandboxMode, &orgsv1.ProvisionRequest{
		OrgId:       org.ID,
		Region:      defaultOrgRegion,
		Reprovision: false,
		CustomCert:  false,
	})
	if err != nil {
		w.mw.Event(ctx, &statsd.Event{
			Title: "org failed to provision",
			Text: fmt.Sprintf(
				"org %s failed to provision\ncreated by %s\nerror: %s",
				org.ID,
				org.CreatedBy.Email,
				err.Error(),
			),
			Tags: metrics.ToTags(map[string]string{
				"status":             "error",
				"status_description": "failed to provision",
			}),
		})
		w.updateStatus(ctx, org.ID, app.OrgStatusError, "unable to provision organization resources")
		return fmt.Errorf("unable to provision org: %w", err)
	}
	w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
		OrgID:       org.ID,
		SandboxMode: sandboxMode,
	})
	w.updateStatus(ctx, org.ID, app.OrgStatusActive, "organization resources are provisioned")
	return nil
}
