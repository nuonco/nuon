package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

// Commented out polling logic - validation now happens at endpoint level
// func (w *Workflows) pollAppsDeprovisioned(ctx workflow.Context, orgID string) error {
// 	for {
// 		org, err := activities.AwaitGetByOrgID(ctx, orgID)
// 		if err != nil {
// 			w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
// 			return fmt.Errorf("unable to get org: %w", err)
// 		}
//
// 		if len(org.Apps) < 1 {
// 			return nil
// 		}
// 		workflow.Sleep(ctx, defaultPollTimeout)
// 	}
// }

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.OrgStatusActive, "checking that all apps are deleted before deprovisioning")

	// Check if any apps still exist - return non-retryable error if so
	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	if len(org.Apps) > 0 {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "cannot deprovision org with active apps")
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("organization has %d app(s) that must be deleted before deprovisioning", len(org.Apps)),
			"AppsStillPresent",
			nil,
		)
	}

	return w.deprovisionOrg(ctx, sreq.ID, sreq.SandboxMode)
}

func (w *Workflows) deprovisionOrg(ctx workflow.Context, orgID string, sandboxMode bool) error {
	l := workflow.GetLogger(ctx)

	org, err := activities.AwaitGet(ctx, activities.GetRequest{
		OrgID: orgID,
	})
	if err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	w.updateStatus(ctx, orgID, app.OrgStatusDeprovisioning, "deprovisioning organization resources")

	// reprovision IAM roles for the org
	orgIAMReq := &orgiam.DeprovisionIAMRequest{
		OrgID: orgID,
	}
	if org.OrgType == app.OrgTypeDefault {
		_, err = orgiam.AwaitDeprovisionIAM(ctx, orgIAMReq)
		if err != nil {
			w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to deprovision iam roles")
			return fmt.Errorf("unable to deprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await deprovision iam",
			zap.Any("org_type", org.OrgType),
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name))
	}

	if len(org.RunnerGroup.Runners) < 1 {
		w.updateStatus(ctx, orgID, app.OrgStatusDeprovisioned, "organization successfully deprovisioned")
		return nil
	}

	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDeprovision,
	})
	w.updateStatus(ctx, orgID, app.OrgStatusDeprovisioned, "organization successfully deprovisioned")
	return nil
}
