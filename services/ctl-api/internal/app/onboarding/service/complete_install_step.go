package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/appconfigupdated"
	installscreated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/polldependencies"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type CompleteInstallStepRequest struct {
	Name        string                    `json:"name" validate:"required,entity_name"`
	InstallMode app.OnboardingInstallMode `json:"install_mode,omitempty" validate:"omitempty,oneof=cloud sandbox" swaggertype:"string" enums:"cloud,sandbox"`

	AWSAccount *struct {
		Region string `json:"region"`
	} `json:"aws_account,omitempty"`

	AzureAccount *struct {
		Location string `json:"location"`
	} `json:"azure_account,omitempty"`

	Inputs map[string]*string `json:"inputs,omitempty"`

	Metadata *struct {
		ManagedBy string `json:"managed_by,omitempty"`
	} `json:"metadata,omitempty"`
}

// @ID						CompleteOnboardingInstallStep
// @Summary				Complete the install step
// @Description			Creates an install and advances the onboarding to the install status step
// @Tags					onboarding
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					req	body	CompleteInstallStepRequest	true	"Input"
// @Success				200	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current/steps/install [POST]
func (s *service) CompleteInstallStep(ctx *gin.Context) {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	onboarding, err := s.getActiveOnboarding(ctx, account.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if onboarding.CurrentStep != app.OnboardingStepInstall {
		ctx.Error(fmt.Errorf("expected step %s but current step is %s", app.OnboardingStepInstall, onboarding.CurrentStep))
		return
	}

	var req CompleteInstallStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Store install mode preference
	if req.InstallMode != "" {
		onboarding.InstallMode = req.InstallMode
	}

	// Validate onboarding has required references
	if onboarding.AppID == nil || *onboarding.AppID == "" {
		ctx.Error(fmt.Errorf("onboarding has no app_id set; cannot create install"))
		return
	}
	if onboarding.OrgID == nil || *onboarding.OrgID == "" {
		ctx.Error(fmt.Errorf("onboarding has no org_id set; cannot create install"))
		return
	}

	// Idempotency: if install already created, advance step and return
	if onboarding.InstallID != nil && *onboarding.InstallID != "" {
		onboarding.CurrentStep = app.OnboardingStepDeploy
		onboarding.StepStatus = app.OnboardingStepStatusActive
		onboarding.SetCompositeStatus(ctx, app.Status(app.OnboardingStepStatusActive))
		if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
			ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
			return
		}
		ctx.JSON(http.StatusOK, onboarding)
		return
	}

	// Auto-populate cloud account if not provided, using the app's cloud platform
	if req.AWSAccount == nil && req.AzureAccount == nil {
		cloudPlatform := s.resolveCloudPlatform(ctx, *onboarding.AppID, onboarding.CloudProvider)
		switch cloudPlatform {
		case app.CloudPlatformAWS:
			req.AWSAccount = &struct {
				Region string `json:"region"`
			}{Region: "us-east-1"}
		case app.CloudPlatformAzure:
			req.AzureAccount = &struct {
				Location string `json:"location"`
			}{Location: "eastus"}
		}
	}

	// Determine sandbox mode
	var org app.Org
	if err := s.db.WithContext(ctx).First(&org, "id = ?", *onboarding.OrgID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}
	sandboxMode := s.cfg.ForceOnboardingSandboxMode || org.SandboxMode || onboarding.InstallMode == app.OnboardingInstallModeSandbox

	// Append a shortcode to make the install name unique
	shortcode, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 6)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to generate shortcode: %w", err))
		return
	}
	installName := fmt.Sprintf("%s-%s", req.Name, shortcode)

	// Build install params
	installParams := &helpers.CreateInstallParams{
		Name:        installName,
		Inputs:      req.Inputs,
		SandboxMode: sandboxMode,
		InstallConfig: &helpers.CreateInstallConfigParams{
			ApprovalOption: app.InstallApprovalOptionApproveAll,
		},
	}
	if req.AWSAccount != nil {
		installParams.AWSAccount = &struct {
			Region string `json:"region"`
		}{Region: req.AWSAccount.Region}
	}
	if req.AzureAccount != nil {
		installParams.AzureAccount = &struct {
			Location string `json:"location"`
		}{Location: req.AzureAccount.Location}
	}
	if req.Metadata != nil {
		installParams.Metadata = helpers.InstallMetadata{
			ManagedBy: req.Metadata.ManagedBy,
		}
	}

	// Create install synchronously
	install, err := s.installsHelpers.CreateInstall(ctx, *onboarding.AppID, installParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	// Create provision workflow
	workflow, err := s.installsHelpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		map[string]string{},
		false,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create provision workflow: %w", err))
		return
	}

	// Send signals: v2 queues or legacy event loop (matching CreateInstallV2 pattern)
	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &installscreated.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &polldependencies.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		if err := s.enqueueInstallSignal(ctx, workflowsQueueID, &executeflow.Signal{
			WorkflowID: workflow.ID,
		}, workflow.ID, "install_workflows"); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		// reconcile cron/drift emitters from app config triggers
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &appconfigupdated.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue reconcile-emitters signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationCreated,
		})
		s.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationPollDependencies,
		})
		s.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type:              installsignals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
		s.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationSyncActionWorkflowTriggers,
		})
	}

	// Update onboarding with install/workflow references and advance step
	onboarding.InstallID = &install.ID
	onboarding.WorkflowID = &workflow.ID
	onboarding.CurrentStep = app.OnboardingStepDeploy
	onboarding.StepStatus = app.OnboardingStepStatusActive
	onboarding.StepError = nil
	onboarding.SetCompositeStatus(ctx, app.Status(app.OnboardingStepStatusActive))

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
		return
	}

	// Update user journey (non-blocking)
	if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, account.ID, install.ID); err != nil {
		s.l.Warn("failed to update user journey for first install create", zap.Error(err))
	}

	ctx.JSON(http.StatusOK, onboarding)
}

// resolveCloudPlatform determines the cloud platform for the app, first checking
// the onboarding's CloudProvider, then falling back to the app's runner config.
func (s *service) resolveCloudPlatform(ctx context.Context, appID string, cloudProvider *string) app.CloudPlatform {
	// Strategy 1: Use the onboarding's CloudProvider if available
	if cloudProvider != nil && *cloudProvider != "" {
		switch *cloudProvider {
		case "aws":
			return app.CloudPlatformAWS
		case "azure":
			return app.CloudPlatformAzure
		case "gcp":
			return app.CloudPlatformGCP
		}
	}

	// Strategy 2: Load from app's runner config
	var runnerConfig app.AppRunnerConfig
	res := s.db.WithContext(ctx).
		Where("app_id = ?", appID).
		Order("created_at DESC").
		First(&runnerConfig)
	if res.Error != nil {
		return app.CloudPlatformUnknown
	}

	// AfterQuery hook on AppRunnerConfig sets CloudPlatform automatically
	return runnerConfig.CloudPlatform
}
