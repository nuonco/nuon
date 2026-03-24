package createapp

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	if err := s.executeCreateApp(ctx, logger); err != nil {
		logger.Error("create app failed", "error", err)

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

func (s *Signal) executeCreateApp(ctx workflow.Context, logger interface{ Info(string, ...interface{}) }) error {
	// Fetch onboarding
	onboarding, err := activities.AwaitGetOnboardingByOnboardingID(ctx, s.OnboardingID)
	if err != nil {
		return fmt.Errorf("unable to get onboarding: %w", err)
	}

	if onboarding.OrgID == nil || *onboarding.OrgID == "" {
		return fmt.Errorf("onboarding has no org_id set")
	}

	// Set org + account context on workflow so the cctx propagator passes it to all activities.
	// Both are required — InjectFromWorkflow fails silently if either is missing.
	ctx = cctx.SetOrgIDWorkflowContext(ctx, *onboarding.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, onboarding.AccountID)

	// Derive app name from example slug or default
	appName := "onboarding-app"
	if onboarding.ExampleAppSlug != nil && *onboarding.ExampleAppSlug != "" {
		appName = *onboarding.ExampleAppSlug
	}

	// Create app + branch (activity handles idempotency — returns existing app if already created)
	appResp, err := activities.AwaitCreateOnboardingApp(ctx, activities.CreateOnboardingAppRequest{
		OrgID:   *onboarding.OrgID,
		AppName: appName,
	})
	if err != nil {
		return fmt.Errorf("unable to create app: %w", err)
	}

	logger.Info("created onboarding app", "app_id", appResp.AppID, "app_branch_id", appResp.AppBranchID)

	// For example apps, create branch config with public git VCS and trigger a sync run
	if s.ExampleRepo != "" {
		branchConfig, err := activities.AwaitCreateOnboardingAppBranchConfig(ctx, activities.CreateOnboardingAppBranchConfigRequest{
			AppBranchID: appResp.AppBranchID,
			Repo:        s.ExampleRepo,
			Directory:   s.ExampleDirectory,
			Branch:      s.ExampleBranch,
		})
		if err != nil {
			return fmt.Errorf("unable to create branch config: %w", err)
		}

		// Trigger app branch run to clone repo, parse config, create components, and build
		runResp, err := activities.AwaitTriggerOnboardingAppBranchRun(ctx, activities.TriggerOnboardingAppBranchRunRequest{
			AppBranchID:       appResp.AppBranchID,
			AppBranchConfigID: branchConfig.ID,
		})
		if err != nil {
			return fmt.Errorf("unable to trigger app branch run: %w", err)
		}

		logger.Info("triggered app branch run", "run_id", runResp.RunID, "workflow_id", runResp.WorkflowID)
	}

	// Update onboarding with app references and advance step
	nextStep := string(app.OnboardingStepInstall)
	stepStatus := string(app.OnboardingStepStatusIdle)
	_, err = activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
		Req: &activities.UpdateOnboardingInput{
			OnboardingID: s.OnboardingID,
			AppID:        &appResp.AppID,
			AppBranchID:  &appResp.AppBranchID,
			CurrentStep:  &nextStep,
			StepStatus:   &stepStatus,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update onboarding with app references: %w", err)
	}

	return nil
}
