package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createorg "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/create_org"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CompleteOrganizationStepRequest struct {
	Name string `json:"name" validate:"required"`
}

// @ID						CompleteOnboardingOrganizationStep
// @Summary				Complete the organization step
// @Description			Creates a sandbox organization and advances the onboarding to the app profile step
// @Tags					onboarding
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Param					req	body	CompleteOrganizationStepRequest	true	"Input"
// @Success				200	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current/steps/organization [POST]
func (s *service) CompleteOrganizationStep(ctx *gin.Context) {
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

	if onboarding.CurrentStep != app.OnboardingStepOrganization {
		ctx.Error(fmt.Errorf("expected step %s but current step is %s", app.OnboardingStepOrganization, onboarding.CurrentStep))
		return
	}

	var req CompleteOrganizationStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Set processing status — signal will advance CurrentStep on completion
	onboarding.StepStatus = app.OnboardingStepStatusProcessing

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

	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &createorg.Signal{
			OnboardingID: onboarding.ID,
			AccountID:    account.ID,
			OrgName:      req.Name,
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue create org signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}
