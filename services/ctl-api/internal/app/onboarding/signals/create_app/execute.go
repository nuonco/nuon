package createapp

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/queuebuild"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	if err := s.executeCreateApp(ctx, logger); err != nil {
		logger.Error("create app failed", "error", err)

		errMsg := signal.HumanError(err)
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
	onboarding, err := activities.AwaitGetOnboardingByOnboardingID(ctx, s.OnboardingID)
	if err != nil {
		return fmt.Errorf("unable to get onboarding: %w", err)
	}

	if onboarding.OrgID == nil || *onboarding.OrgID == "" {
		return fmt.Errorf("onboarding has no org_id set")
	}

	// why: Set org + account context on workflow so the cctx propagator passes it to all activities.
	// Both are required — InjectFromWorkflow fails silently if either is missing.
	ctx = cctx.SetOrgIDWorkflowContext(ctx, *onboarding.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, onboarding.AccountID)

	appName := "onboarding-app"
	if onboarding.ExampleAppSlug != nil && *onboarding.ExampleAppSlug != "" {
		appName = *onboarding.ExampleAppSlug
	}

	isExampleApp := s.ExampleRepo != ""

	appResp, err := activities.AwaitCreateOnboardingApp(ctx, activities.CreateOnboardingAppRequest{
		OrgID:        *onboarding.OrgID,
		AppName:      appName,
		CreateBranch: isExampleApp,
	})
	if err != nil {
		return fmt.Errorf("unable to create app: %w", err)
	}

	logger.Info("created onboarding app", "app_id", appResp.AppID, "app_branch_id", appResp.AppBranchID)

	updateReq := &activities.UpdateOnboardingInput{
		OnboardingID: s.OnboardingID,
		AppID:        &appResp.AppID,
	}
	if appResp.AppBranchID != "" {
		updateReq.AppBranchID = &appResp.AppBranchID
	}
	_, err = activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
		Req: updateReq,
	})
	if err != nil {
		return fmt.Errorf("unable to save app references on onboarding: %w", err)
	}

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

		cb := callback.New(ctx, s.OnboardingID)
		runResp, err := activities.AwaitTriggerOnboardingAppBranchRun(ctx, activities.TriggerOnboardingAppBranchRunRequest{
			AppBranchID:       appResp.AppBranchID,
			AppBranchConfigID: branchConfig.ID,
			Cb:                cb,
		})
		if err != nil {
			return fmt.Errorf("unable to trigger app branch run: %w", err)
		}

		logger.Info("triggered app branch run", "run_id", runResp.RunID, "workflow_id", runResp.WorkflowID)

		_, err = callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout)
		if err != nil {
			return fmt.Errorf("app branch run failed: %w", err)
		}

		logger.Info("app branch run completed", "run_id", runResp.RunID)

		nextStep := string(app.OnboardingStepInstall)
		stepStatus := string(app.OnboardingStepStatusActive)
		_, err = activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
			Req: &activities.UpdateOnboardingInput{
				OnboardingID: s.OnboardingID,
				CurrentStep:  &nextStep,
				StepStatus:   &stepStatus,
			},
		})
		if err != nil {
			return fmt.Errorf("unable to advance onboarding step: %w", err)
		}
	} else if onboarding.CloudProvider != nil && *onboarding.CloudProvider != "" {
		_, err := activities.AwaitBuildCustomAppConfig(ctx, activities.BuildCustomAppConfigRequest{
			OnboardingID: s.OnboardingID,
		})
		if err != nil {
			return fmt.Errorf("unable to build custom app config: %w", err)
		}

		syncResp, err := activities.AwaitSyncCustomAppConfig(ctx, activities.SyncCustomAppConfigRequest{
			OnboardingID: s.OnboardingID,
		})
		if err != nil {
			return fmt.Errorf("unable to sync custom app config: %w", err)
		}

		logger.Info("synced custom app config", "app_config_id", syncResp.AppConfigID, "component_ids", syncResp.ComponentIDs)

		nextStep := string(app.OnboardingStepInstall)
		stepStatus := string(app.OnboardingStepStatusActive)
		_, err = activities.AwaitUpdateOnboarding(ctx, activities.UpdateOnboardingRequest{
			Req: &activities.UpdateOnboardingInput{
				OnboardingID: s.OnboardingID,
				CurrentStep:  &nextStep,
				StepStatus:   &stepStatus,
			},
		})
		if err != nil {
			return fmt.Errorf("unable to advance onboarding step: %w", err)
		}

		for _, componentID := range syncResp.ComponentIDs {
			// why: Don't pass AppConfigID for custom apps — there's no AppBranchRun,
			// and the queuebuild signal would retry GetAppBranchRunByAppConfigID forever.
			_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
				OwnerID:   componentID,
				OwnerType: "components",
				QueueName: queuenames.ComponentDefaultQueueName,
				Signal: &queuebuild.Signal{
					ComponentID: componentID,
				},
			})
			if err != nil {
				return fmt.Errorf("component %s: enqueue build failed: %w", componentID, err)
			}

			logger.Info("enqueued component build", "component_id", componentID)
		}
	}

	return nil
}
