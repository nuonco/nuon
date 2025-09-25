package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateUserJourneyRequest struct {
	Name  string                     `json:"name" binding:"required"`
	Title string                     `json:"title" binding:"required"`
	Steps []CreateUserJourneyStepReq `json:"steps" binding:"required"`
}

type CreateUserJourneyStepReq struct {
	Name  string `json:"name" binding:"required"`
	Title string `json:"title" binding:"required"`
}

// @ID						CreateUserJourney
// @Summary				Create a new user journey for account
// @Description			Add a new user journey with steps to track user progress
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Param					body	body		CreateUserJourneyRequest	true	"Create journey request"
// @Failure				400		{object}	stderr.ErrResponse
// @Failure				401		{object}	stderr.ErrResponse
// @Failure				403		{object}	stderr.ErrResponse
// @Failure				404		{object}	stderr.ErrResponse
// @Failure				500		{object}	stderr.ErrResponse
// @Success				201		{object}	app.Account
// @Router					/v1/account/user-journeys [POST]
func (s *service) CreateUserJourney(ctx *gin.Context) {
	var req CreateUserJourneyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Delegate business logic to private method
	updatedAccount, err := s.createUserJourney(ctx, account.ID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, updatedAccount)
}

func (s *service) createUserJourney(ctx *gin.Context, accountID string, req *CreateUserJourneyRequest) (*app.Account, error) {
	account, err := s.getAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Check for duplicate journey names
	for _, journey := range account.UserJourneys {
		if journey.Name == req.Name {
			return nil, fmt.Errorf("journey with name '%s' already exists", req.Name)
		}
	}

	// Create journey steps
	steps := make([]app.UserJourneyStep, len(req.Steps))
	for i, stepReq := range req.Steps {
		steps[i] = app.UserJourneyStep{
			Name:     stepReq.Name,
			Title:    stepReq.Title,
			Complete: false,
		}
	}

	// Create new journey
	newJourney := app.UserJourney{
		Name:  req.Name,
		Title: req.Title,
		Steps: steps,
	}

	// Add journey to account
	if account.UserJourneys == nil {
		account.UserJourneys = []app.UserJourney{}
	}
	account.UserJourneys = append(account.UserJourneys, newJourney)

	// Save to database
	if err := s.db.WithContext(ctx).Save(account).Error; err != nil {
		return nil, fmt.Errorf("unable to create user journey: %w", err)
	}

	return account, nil
}
