package provision

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
	runnerprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/provision"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "org-provision"

const (
	pollRunnerTimeout time.Duration = time.Minute * 5
	pollRunnerPeriod  time.Duration = time.Second * 10
)

type Signal struct {
	signal.Hooks
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

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

	s.updateStatus(ctx, app.OrgStatusProvisioning, "provisioning organization resources")

	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	// Provision IAM roles for the org
	if org.OrgType == app.OrgTypeDefault {
		orgIAMReq := &orgiam.ProvisionIAMRequest{
			OrgID:      s.OrgID,
			RunnerID:   org.RunnerGroup.Runners[0].ID,
			WorkflowID: fmt.Sprintf("%s-provision-iam", workflow.GetInfo(ctx).WorkflowExecution.ID),
		}
		iamResp, err := orgiam.AwaitProvisionIAM(ctx, orgIAMReq)
		if err != nil {
			s.updateStatus(ctx, app.OrgStatusError, "unable to provision IAM")
			return fmt.Errorf("unable to provision IAM: %w", err)
		}

		// Persist per-org Azure client ID to runner group settings so that
		// runner provisioning and reprovision signals can read it back.
		if iamResp.AzureClientID != "" {
			if err := activities.AwaitUpdateRunnerGroupAzureClientID(ctx, activities.UpdateRunnerGroupAzureClientIDRequest{
				OrgID:         s.OrgID,
				AzureClientID: iamResp.AzureClientID,
			}); err != nil {
				s.updateStatus(ctx, app.OrgStatusError, "unable to update runner group azure client ID")
				return fmt.Errorf("unable to update runner group azure client ID: %w", err)
			}
		}
	} else {
		l.Info("skipping await provision iam",
			zap.Any("org_type", org.OrgType),
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name))
	}

	// Provision the runner via v2 queue signal
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   org.RunnerGroup.Runners[0].ID,
		OwnerType: "runners",
		Signal: &runnerprovision.Signal{
			RunnerID: org.RunnerGroup.Runners[0].ID,
		},
	})
	if err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "unable to enqueue runner provision signal")
		return fmt.Errorf("unable to enqueue runner provision signal: %w", err)
	}

	if err := s.pollRunner(ctx, org.RunnerGroup.Runners[0].ID); err != nil {
		s.updateStatus(ctx, app.OrgStatusError, "organization did not provision runner")
		return fmt.Errorf("runner did not provision correctly: %w", err)
	}

	s.updateStatus(ctx, app.OrgStatusActive, "organization resources are provisioned")
	return nil
}

func (s *Signal) updateStatus(ctx workflow.Context, status app.OrgStatus, statusDescription string) {
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		OrgID:             s.OrgID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Error("unable to update org status",
			zap.String("organization-id", s.OrgID),
			zap.Error(err))
	}
	if err := statusactivities.AwaitUpdateOrgStatusV2(ctx, statusactivities.UpdateOrgStatusV2Request{
		OrgID:             s.OrgID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Error("unable to update org status v2",
			zap.String("organization-id", s.OrgID),
			zap.Error(err))
	}
}

func (s *Signal) pollRunner(ctx workflow.Context, runnerID string) error {
	timeout := workflow.Now(ctx).Add(pollRunnerTimeout)

	var lastStatus app.RunnerStatus
	for !workflow.Now(ctx).After(timeout) {
		runner, err := activities.AwaitGetRunner(ctx, activities.GetRunnerRequest{
			ID: runnerID,
		})
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
