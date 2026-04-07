package delete

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam"
	runnerdelete "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/delete"
	runnerdeprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/deprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "org-delete"

const (
	pollRunnerPeriod = 10 * time.Second
)

type Signal struct {
	OrgID       string `json:"org_id"`
	ForceDelete bool   `json:"force_delete"`
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
	l := workflow.GetLogger(ctx)

	s.updateStatus(ctx, app.OrgStatusDeleting, "ensuring all apps are deleted before deprovisioning")

	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	// Skip deprovision if org is already deprovisioned
	if org.Status != app.OrgStatusDeprovisioned {
		err = s.deprovision(ctx)
		if err != nil {
			if !s.ForceDelete {
				return err
			}
			l.Error("unable to deprovision org, continuing anyway", zap.Error(err))
		}
	} else {
		l.Info("skipping deprovision, org already deprovisioned", zap.String("org_id", org.ID))
	}

	if len(org.RunnerGroup.Runners) > 0 {
		_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   org.RunnerGroup.Runners[0].ID,
			OwnerType: "runners",
			Signal: &runnerdelete.Signal{
				RunnerID: org.RunnerGroup.Runners[0].ID,
			},
		})
		if err != nil {
			if !s.ForceDelete {
				return fmt.Errorf("unable to enqueue runner delete signal: %w", err)
			}
			l.Error("unable to enqueue runner delete signal, continuing anyway", zap.Error(err))
		}

		err = s.pollRunnerNotFound(ctx, org.RunnerGroup.Runners[0].ID)
		if err != nil {
			if !s.ForceDelete {
				return fmt.Errorf("unable to poll runner to not found: %w", err)
			}
			l.Error("unable to poll runner to not found, continuing anyway", zap.Error(err))
		}
	}

	if err := activities.AwaitDeleteByOrgID(ctx, s.OrgID); err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to delete organization from database")
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}

func (s *Signal) deprovision(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Check for apps
	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}
	if len(org.Apps) > 0 {
		s.updateStatus(ctx, app.OrgStatusError, "cannot deprovision org with active apps")
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("organization has %d app(s) that must be deleted before deprovisioning", len(org.Apps)),
			"AppsStillPresent",
			nil,
		)
	}

	s.updateStatus(ctx, app.OrgStatusDeprovisioning, "deprovisioning organization resources")

	orgFull, err := activities.AwaitGet(ctx, activities.GetRequest{OrgID: s.OrgID})
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	if orgFull.OrgType == app.OrgTypeDefault {
		_, err = orgiam.AwaitDeprovisionIAM(ctx, &orgiam.DeprovisionIAMRequest{OrgID: s.OrgID, WorkflowID: fmt.Sprintf("%s-deprovision-iam", workflow.GetInfo(ctx).WorkflowExecution.ID)})
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to deprovision iam roles")
			return fmt.Errorf("unable to deprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await deprovision iam", zap.Any("org_type", orgFull.OrgType))
	}

	if len(orgFull.RunnerGroup.Runners) > 0 {
		_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   orgFull.RunnerGroup.Runners[0].ID,
			OwnerType: "runners",
			Signal: &runnerdeprovision.Signal{
				RunnerID: orgFull.RunnerGroup.Runners[0].ID,
			},
		})
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to enqueue runner deprovision signal")
			return fmt.Errorf("unable to enqueue runner deprovision signal: %w", err)
		}
	}
	s.updateStatus(ctx, app.OrgStatusDeprovisioned, "organization successfully deprovisioned")
	return nil
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

func (s *Signal) pollRunnerNotFound(ctx workflow.Context, runnerID string) error {
	for {
		runner, err := activities.AwaitGetRunner(ctx, activities.GetRunnerRequest{ID: runnerID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("unable to get runner from database: %w", err)
		}
		_ = runner
		workflow.Sleep(ctx, pollRunnerPeriod)
	}
}
