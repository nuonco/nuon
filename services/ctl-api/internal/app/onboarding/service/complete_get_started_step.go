package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						CompleteOnboardingGetStartedStep
// @Summary				Complete onboarding
// @Description			Marks the onboarding session as completed
// @Tags					onboarding
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Success				200	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current/steps/get-started [POST]
func (s *service) CompleteGetStartedStep(ctx *gin.Context) {
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

	if onboarding.CurrentStep != app.OnboardingStepGetStarted {
		ctx.Error(fmt.Errorf("expected step %s but current step is %s", app.OnboardingStepGetStarted, onboarding.CurrentStep))
		return
	}

	onboarding.Status = app.OnboardingStatusCompleted
	onboarding.SetCompositeStatus(ctx, app.Status(app.OnboardingStatusCompleted))

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}
