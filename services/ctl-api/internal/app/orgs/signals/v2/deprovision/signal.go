package deprovision

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appdeprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/deprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam"
	runnerdeprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/deprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "org-deprovision"

type Signal struct {
	OrgID string `json:"org_id"`
	Force bool   `json:"force"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	_, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	if len(org.Apps) > 0 {
		if !s.Force {
			s.updateStatus(ctx, app.OrgStatusError, "cannot deprovision org with active apps")
			return temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("organization has %d app(s) that must be deleted before deprovisioning", len(org.Apps)),
				"AppsStillPresent",
				nil,
			)
		}

		// Force mode: deprovision and delete all apps first
		l := workflow.GetLogger(ctx)
		s.updateStatus(ctx, app.OrgStatusDeprovisioning, "force deprovisioning: deleting all apps")
		for _, a := range org.Apps {
			l.Info("enqueuing app deprovision signal", zap.String("app_id", a.ID))
			_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
				OwnerID:   a.ID,
				OwnerType: "apps",
				Signal: &appdeprovision.Signal{
					AppID: a.ID,
				},
			})
			if err != nil {
				l.Error("unable to enqueue app deprovision signal, continuing anyway", zap.String("app_id", a.ID), zap.Error(err))
			}
		}
	}

	return s.deprovisionOrg(ctx)
}

func (s *Signal) deprovisionOrg(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	org, err := activities.AwaitGet(ctx, activities.GetRequest{OrgID: s.OrgID})
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	s.updateStatus(ctx, app.OrgStatusDeprovisioning, "deprovisioning organization resources")

	if org.OrgType == app.OrgTypeDefault {
		_, err = orgiam.AwaitDeprovisionIAM(ctx, &orgiam.DeprovisionIAMRequest{OrgID: s.OrgID, WorkflowID: fmt.Sprintf("%s-deprovision-iam", workflow.GetInfo(ctx).WorkflowExecution.ID)})
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to deprovision iam roles")
			return fmt.Errorf("unable to deprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await deprovision iam", zap.Any("org_type", org.OrgType), zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	}

	if len(org.RunnerGroup.Runners) < 1 {
		s.updateStatus(ctx, app.OrgStatusDeprovisioned, "organization successfully deprovisioned")
		return nil
	}

	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   org.RunnerGroup.Runners[0].ID,
		OwnerType: "runners",
		Signal: &runnerdeprovision.Signal{
			RunnerID: org.RunnerGroup.Runners[0].ID,
		},
	})
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to enqueue runner deprovision signal")
		return fmt.Errorf("unable to enqueue runner deprovision signal: %w", err)
	}
	s.updateStatus(ctx, app.OrgStatusDeprovisioned, "organization successfully deprovisioned")
	return nil
}

func (s *Signal) updateStatus(ctx workflow.Context, status app.OrgStatus, statusDescription string) {
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{OrgID: s.OrgID, Status: status, StatusDescription: statusDescription}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Error("unable to update org status", zap.String("organization-id", s.OrgID), zap.Error(err))
	}
}
