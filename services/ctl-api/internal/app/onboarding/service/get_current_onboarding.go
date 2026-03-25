package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetCurrentOnboarding
// @Summary				Get current onboarding session
// @Description			Returns the active onboarding session for the current account
// @Tags					onboarding
// @Produce				json
// @Security				APIKey
// @Success				200	{object}	app.Onboarding
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current [GET]
func (s *service) GetCurrentOnboarding(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, onboarding)
}

// getActiveOnboarding loads the current active onboarding for an account.
func (s *service) getActiveOnboarding(ctx *gin.Context, accountID string) (*app.Onboarding, error) {
	var onboarding app.Onboarding
	res := s.db.WithContext(ctx).
		Where("account_id = ? AND status = ?", accountID, app.OnboardingStatusActive).
		Order("created_at DESC").
		First(&onboarding)
	if res.Error != nil {
		return nil, fmt.Errorf("no active onboarding session found: %w", res.Error)
	}
	return &onboarding, nil
}
