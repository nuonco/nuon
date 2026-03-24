package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						CompleteOnboardingDeployStep
// @Summary				Complete the deploy step
// @Description			Advances the onboarding past the deploy monitoring to get started
// @Tags					onboarding
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Success				200	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current/steps/deploy [POST]
func (s *service) CompleteDeployStep(ctx *gin.Context) {
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

	if onboarding.CurrentStep != app.OnboardingStepDeploy {
		ctx.Error(fmt.Errorf("expected step %s but current step is %s", app.OnboardingStepDeploy, onboarding.CurrentStep))
		return
	}

	onboarding.CurrentStep = app.OnboardingStepGetStarted

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}
