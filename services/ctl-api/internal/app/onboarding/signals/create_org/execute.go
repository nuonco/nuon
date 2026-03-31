package createorg

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	pollOrgTimeout time.Duration = 10 * time.Minute
	pollOrgPeriod  time.Duration = 10 * time.Second
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	if err := s.executeCreateOrg(ctx, logger); err != nil {
		logger.Error("create org failed", "error", err)

		errMsg := err.Error()
		stepStatus := string(app.OnboardingStepStatusError)
		_, updateErr := activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
			Req: &activities.UpdateOnboardingInput{
				OnboardingID: s.OnboardingID,
				StepStatus:   &stepStatus,
				StepError:    &errMsg,
			},
		})
		if updateErr != nil {
			logger.Error("unable to set step status to error", "error", updateErr)
		}

		return err
	}

	return nil
}

func (s *Signal) executeCreateOrg(ctx workflow.Context, logger interface{ Info(string, ...interface{}) }) error {
	// Fetch onboarding
	onboarding, err := activities.AwaitGetOnboardingByOnboardingID(ctx, s.OnboardingID)
	if err != nil {
		return fmt.Errorf("unable to get onboarding: %w", err)
	}

	// Set account context so the cctx propagator passes it to all activities.
	// Required for GORM BeforeCreate hooks that populate CreatedByID.
	ctx = cctx.SetAccountIDWorkflowContext(ctx, onboarding.AccountID)

	// Idempotency: skip if org already created
	if onboarding.OrgID != nil && *onboarding.OrgID != "" {
		logger.Info("org already created, skipping", "org_id", *onboarding.OrgID)
		return nil
	}

	// Create org with sandbox mode
	org, err := activities.AwaitCreateOnboardingOrg(ctx, activities.CreateOnboardingOrgRequest{
		AccountID: onboarding.AccountID,
		OrgName:   s.OrgName,
	})
	if err != nil {
		return fmt.Errorf("unable to create org: %w", err)
	}

	logger.Info("created onboarding org", "org_id", org.ID)

	// Poll org status until provisioned or errored
	if err := s.pollOrgStatus(ctx, org.ID); err != nil {
		return fmt.Errorf("org provisioning failed: %w", err)
	}

	// Update onboarding with org reference and advance step
	nextStep := string(app.OnboardingStepYourStack)
	stepStatus := string(app.OnboardingStepStatusActive)
	_, err = activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
		Req: &activities.UpdateOnboardingInput{
			OnboardingID: s.OnboardingID,
			OrgID:        &org.ID,
			CurrentStep:  &nextStep,
			StepStatus:   &stepStatus,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update onboarding with org reference: %w", err)
	}

	return nil
}

func (s *Signal) pollOrgStatus(ctx workflow.Context, orgID string) error {
	timeout := workflow.Now(ctx).Add(pollOrgTimeout)

	var lastStatus app.OrgStatus
	for !workflow.Now(ctx).After(timeout) {
		org, err := activities.AwaitGetOrgByIDByOrgID(ctx, orgID)
		if err != nil {
			return fmt.Errorf("unable to get org: %w", err)
		}

		switch org.Status {
		case app.OrgStatusActive:
			return nil
		case app.OrgStatusError:
			return fmt.Errorf("org entered error state: %s", org.StatusDescription)
		}

		lastStatus = org.Status
		workflow.Sleep(ctx, pollOrgPeriod)
	}

	return fmt.Errorf("org did not reach active status after %s — last status: %s", pollOrgTimeout, lastStatus)
}
