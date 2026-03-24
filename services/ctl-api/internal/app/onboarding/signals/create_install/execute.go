package createinstall

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	if err := s.executeCreateInstall(ctx, logger); err != nil {
		logger.Error("create install failed", "error", err)

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

func (s *Signal) executeCreateInstall(ctx workflow.Context, logger interface{ Info(string, ...interface{}) }) error {
	// Fetch onboarding
	onboarding, err := activities.AwaitGetOnboardingByOnboardingID(ctx, s.OnboardingID)
	if err != nil {
		return fmt.Errorf("unable to get onboarding: %w", err)
	}

	// Idempotency: skip if install already created
	if onboarding.InstallID != nil && *onboarding.InstallID != "" {
		logger.Info("install already created, skipping", "install_id", *onboarding.InstallID)
		return nil
	}

	if onboarding.AppID == nil || *onboarding.AppID == "" {
		return fmt.Errorf("onboarding has no app_id set; cannot create install")
	}

	if onboarding.OrgID == nil || *onboarding.OrgID == "" {
		return fmt.Errorf("onboarding has no org_id set; cannot create install")
	}

	// Set org + account context on workflow so the cctx propagator passes it to all activities.
	// Both are required — InjectFromWorkflow fails silently if either is missing.
	ctx = cctx.SetOrgIDWorkflowContext(ctx, *onboarding.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, onboarding.AccountID)

	// Create install + workflow
	installResp, err := activities.AwaitCreateOnboardingInstall(ctx, activities.CreateOnboardingInstallRequest{
		Input: &activities.CreateOnboardingInstallInput{
			OnboardingID: s.OnboardingID,
			AppID:        *onboarding.AppID,
			Name:         s.InstallName,
			AWSAccount:   s.AWSAccount,
			AzureAccount: s.AzureAccount,
			Inputs:       s.Inputs,
			Config:       &activities.CreateOnboardingInstallConfig{ApprovalOption: "approve-all"},
			Metadata:     s.Metadata,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create install: %w", err)
	}

	logger.Info("created onboarding install", "install_id", installResp.InstallID, "workflow_id", installResp.WorkflowID)

	// Update onboarding with install references and advance step
	nextStep := string(app.OnboardingStepDeploy)
	stepStatus := string(app.OnboardingStepStatusIdle)
	_, err = activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
		Req: &activities.UpdateOnboardingInput{
			OnboardingID: s.OnboardingID,
			InstallID:    &installResp.InstallID,
			WorkflowID:   &installResp.WorkflowID,
			CurrentStep:  &nextStep,
			StepStatus:   &stepStatus,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update onboarding with install references: %w", err)
	}

	return nil
}
