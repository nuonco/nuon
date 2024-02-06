package worker

import (
	"fmt"
	"time"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollAppsDeprovisioned(ctx workflow.Context, orgID string) error {
	for {
		var org app.Org
		if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
			OrgID: orgID,
		}, &org); err != nil {
			w.updateStatus(ctx, orgID, "error", "unable to get org from database")
			return fmt.Errorf("unable to get org: %w", err)
		}

		if len(org.Apps) < 1 {
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

func (w *Workflows) deprovision(ctx workflow.Context, orgID string, sandboxMode bool) error {
	w.updateStatus(ctx, orgID, StatusActive, "ensuring all apps are deleted before deprovisioning")
	if err := w.pollAppsDeprovisioned(ctx, orgID); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "error polling apps being deprovisioned")
		return fmt.Errorf("unable to poll for deleted apps: %w", err)
	}

	w.updateStatus(ctx, orgID, StatusDeprovisioning, "deprovisioning organization resources")
	_, err := w.execDeprovisionWorkflow(ctx, sandboxMode, &orgsv1.DeprovisionRequest{
		OrgId:  orgID,
		Region: defaultOrgRegion,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to deprovision organization resources")
		return fmt.Errorf("unable to deprovision org: %w", err)
	}

	return nil
}
