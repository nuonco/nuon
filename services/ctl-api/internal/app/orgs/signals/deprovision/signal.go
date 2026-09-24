package deprovision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appdeprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/deprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
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
		// Check if any apps have installs — installs must be forgotten first
		for _, a := range org.Apps {
			if len(a.Installs) > 0 {
				s.updateStatus(ctx, app.OrgStatusError, "cannot deprovision: apps have installs that must be forgotten first")
				return temporal.NewNonRetryableApplicationError(
					fmt.Sprintf("app %s has %d install(s) that must be forgotten before deprovisioning", a.ID, len(a.Installs)),
					"InstallsStillPresent",
					nil,
				)
			}
		}

		// No installs — safe to delete apps as part of org deprovision
		l := workflow.GetLogger(ctx)
		s.updateStatus(ctx, app.OrgStatusDeprovisioning, "deprovisioning: deleting all apps")
		for _, a := range org.Apps {
			l.Info("enqueuing app deprovision signal", zap.String("app_id", a.ID))
			_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
				OwnerID:   a.ID,
				OwnerType: "apps",
				QueueName: queuenames.AppSignalsQueueName,
				Signal: &appdeprovision.Signal{
					AppID: a.ID,
				},
			})
			if err != nil {
				l.Error("unable to enqueue app deprovision signal, continuing anyway", zap.String("app_id", a.ID), zap.Error(err))
			}
		}

		// Wait for all apps to be deleted before proceeding
		if err := s.pollAppsDeleted(ctx); err != nil {
			return err
		}
	}

	return s.deprovisionOrg(ctx)
}

func (s *Signal) deprovisionOrg(ctx workflow.Context) error {
	if _, err := activities.AwaitGet(ctx, activities.GetRequest{OrgID: s.OrgID}); err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	s.updateStatus(ctx, app.OrgStatusDeprovisioning, "deprovisioning organization resources")

	s.updateStatus(ctx, app.OrgStatusDeprovisioned, "organization successfully deprovisioned")
	return nil
}

func (s *Signal) pollAppsDeleted(ctx workflow.Context) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)
	l := workflow.GetLogger(ctx)

	for {
		org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to get org from database")
			return fmt.Errorf("unable to get org: %w", err)
		}

		if len(org.Apps) == 0 {
			l.Info("all apps deleted, proceeding with org deprovision")
			return nil
		}

		if workflow.Now(ctx).After(deadline) {
			s.updateStatus(ctx, app.OrgStatusError, "timeout waiting for apps to be deleted")
			return temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("timeout waiting for %d app(s) to be deleted", len(org.Apps)),
				"AppsDeleteTimeout",
				nil,
			)
		}

		s.updateStatus(ctx, app.OrgStatusDeprovisioning, fmt.Sprintf("waiting for %d app(s) to be deleted", len(org.Apps)))
		workflow.Sleep(ctx, time.Second*10)
	}
}

func (s *Signal) updateStatus(ctx workflow.Context, status app.OrgStatus, statusDescription string) {
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{OrgID: s.OrgID, Status: status, StatusDescription: statusDescription}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Error("unable to update org status", zap.String("organization-id", s.OrgID), zap.Error(err))
	}
	if err := statusactivities.AwaitUpdateOrgStatusV2(ctx, statusactivities.UpdateOrgStatusV2Request{
		OrgID:             s.OrgID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Error("unable to update org status v2", zap.String("organization-id", s.OrgID), zap.Error(err))
	}
}
