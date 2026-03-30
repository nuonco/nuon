package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						CreateOnboarding
// @Summary				Start a new onboarding session
// @Description			Creates a new active onboarding session for the current account
// @Tags					onboarding
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Success				201	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding [POST]
func (s *service) CreateOnboarding(ctx *gin.Context) {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Return existing active session if one exists (idempotent)
	var existing app.Onboarding
	res := s.db.WithContext(ctx).
		Where("account_id = ? AND status = ?", account.ID, app.OnboardingStatusActive).
		First(&existing)
	if res.Error == nil {
		ctx.JSON(http.StatusOK, existing)
		return
	}

	onboarding := app.Onboarding{
		AccountID:   account.ID,
		Status:      app.OnboardingStatusActive,
		CurrentStep: app.OnboardingStepOrganization,
	}

	if err := s.db.WithContext(ctx).Create(&onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create onboarding: %w", err))
		return
	}

	// Create a queue for this onboarding session
	_, err = s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     onboarding.ID,
		OwnerType:   plugins.TableName(s.db, app.Onboarding{}),
		Namespace:   "onboardings",
		MaxInFlight: 1,
		MaxDepth:    10,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create onboarding queue: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, onboarding)
}
