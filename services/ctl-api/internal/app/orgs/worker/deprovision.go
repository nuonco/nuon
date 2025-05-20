package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/iam"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollAppsDeprovisioned(ctx workflow.Context, orgID string) error {
	for {
		org, err := activities.AwaitGetByOrgID(ctx, orgID)
		if err != nil {
			w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
			return fmt.Errorf("unable to get org: %w", err)
		}

		if len(org.Apps) < 1 {
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}
}

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.OrgStatusActive, "ensuring all apps are deleted before deprovisioning")
	if err := w.pollAppsDeprovisioned(ctx, sreq.ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "error polling apps being deprovisioned")
		return fmt.Errorf("unable to poll for deleted apps: %w", err)
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
		return nil
	}

	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDeprovision,
	})
	return nil
}
