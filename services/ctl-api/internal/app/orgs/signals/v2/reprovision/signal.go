package reprovision

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam"
	runnerreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/reprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "org-reprovision"

const (
	pollRunnerTimeout = 5 * time.Minute
	pollRunnerPeriod  = 10 * time.Second
)

type Signal struct {
	signal.Hooks
	OrgID string `json:"org_id"`
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

	s.updateStatus(ctx, app.OrgStatusProvisioning, "reprovisioning organization resources")

	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	// deprovision IAM roles
	if org.OrgType == app.OrgTypeDefault {
		_, err = orgiam.AwaitDeprovisionIAM(ctx, &orgiam.DeprovisionIAMRequest{OrgID: s.OrgID, WorkflowID: fmt.Sprintf("%s-deprovision-iam", workflow.GetInfo(ctx).WorkflowExecution.ID)})
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to deprovision iam roles")
			return fmt.Errorf("unable to deprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await deprovision iam", zap.Any("org_type", org.OrgType), zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	}

	// provision IAM roles
	if org.OrgType == app.OrgTypeDefault {
		_, err = orgiam.AwaitProvisionIAM(ctx, &orgiam.ProvisionIAMRequest{OrgID: s.OrgID, Reprovision: true, WorkflowID: fmt.Sprintf("%s-provision-iam", workflow.GetInfo(ctx).WorkflowExecution.ID)})
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to reprovision iam roles")
			return fmt.Errorf("unable to reprovision iam roles: %w", err)
		}
	} else {
		l.Info("skipping await reprovision iam", zap.Any("org_type", org.OrgType), zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	}

	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   org.RunnerGroup.Runners[0].ID,
		OwnerType: "runners",
		Signal: &runnerreprovision.Signal{
			RunnerID: org.RunnerGroup.Runners[0].ID,
		},
	})
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to enqueue runner reprovision signal")
		return fmt.Errorf("unable to enqueue runner reprovision signal: %w", err)
	}

	if err := s.pollRunner(ctx, org.RunnerGroup.Runners[0].ID); err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "organization did not provision runner")
		return fmt.Errorf("runner did not reprovision correctly: %w", err)
	}

	s.updateStatus(ctx, app.OrgStatusActive, "organization resources are provisioned")
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

func (s *Signal) pollRunner(ctx workflow.Context, runnerID string) error {
	timeout := workflow.Now(ctx).Add(pollRunnerTimeout)
	var lastStatus app.RunnerStatus
	for !workflow.Now(ctx).After(timeout) {
		runner, err := activities.AwaitGetRunner(ctx, activities.GetRunnerRequest{ID: runnerID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("unable to get runner from database: %w", err)
		}
		if runner.Status == app.RunnerStatusActive {
			return nil
		}
		lastStatus = runner.Status
		workflow.Sleep(ctx, pollRunnerPeriod)
	}
	return fmt.Errorf("runner did not reach status after %s - last status %s", pollRunnerTimeout, lastStatus)
}
