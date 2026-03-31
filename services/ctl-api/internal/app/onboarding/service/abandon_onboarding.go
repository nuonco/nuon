package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						AbandonOnboarding
// @Summary				Abandon onboarding session
// @Description			Marks the current active onboarding session as abandoned
// @Tags					onboarding
// @Produce				json
// @Security				APIKey
// @Success				200	{object}	app.Onboarding
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current [DELETE]
func (s *service) AbandonOnboarding(ctx *gin.Context) {
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

	onboarding.Status = app.OnboardingStatusAbandoned
	onboarding.SetCompositeStatus(ctx, app.Status(app.OnboardingStatusAbandoned))

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to abandon onboarding: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}
