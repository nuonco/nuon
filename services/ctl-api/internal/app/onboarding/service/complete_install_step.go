package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	createinstall "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/create_install"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CompleteInstallStepRequest struct {
	Name        string                    `json:"name" validate:"required"`
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

	// Advance step immediately so the frontend can proceed; signal will
	// set the same value on completion (idempotent) plus clear StepStatus.
	onboarding.CurrentStep = app.OnboardingStepDeploy
	onboarding.StepStatus = app.OnboardingStepStatusProcessing
	onboarding.SetCompositeStatus(ctx, app.Status(app.OnboardingStepStatusProcessing))

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
		return
	}

	// Get queue and enqueue signal
	queue, err := s.queueClient.GetQueueByOwner(ctx, onboarding.ID, plugins.TableName(s.db, app.Onboarding{}))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get onboarding queue: %w", err))
		return
	}

	signal := &createinstall.Signal{
		OnboardingID: onboarding.ID,
		InstallName:  req.Name,
		Inputs:       req.Inputs,
	}
	if req.AWSAccount != nil {
		signal.AWSAccount = &activities.CreateOnboardingInstallAWS{
			Region: req.AWSAccount.Region,
		}
	}
	if req.AzureAccount != nil {
		signal.AzureAccount = &activities.CreateOnboardingInstallAzure{
			Location: req.AzureAccount.Location,
		}
	}
	if req.Metadata != nil {
		signal.Metadata = &activities.CreateOnboardingInstallMetadata{
			ManagedBy: req.Metadata.ManagedBy,
		}
	}

	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  signal,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue create install signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}
