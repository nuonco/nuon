package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateUserJourneyRequest struct {
	Name  string                      `json:"name" binding:"required"`
	Title string                      `json:"title" binding:"required"`
	Steps []CreateUserJourneyStepReq `json:"steps" binding:"required"`
}

type CreateUserJourneyStepReq struct {
	Name  string `json:"name" binding:"required"`
	Title string `json:"title" binding:"required"`
}

//	@ID						CreateOrgUserJourney
//	@Summary				Create a new user journey for organization
//	@Description			Add a new user journey with steps to track user progress
//	@Tags					orgs
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Param					body	body		CreateUserJourneyRequest	true	"Create journey request"
//	@Failure				400		{object}	stderr.ErrResponse
//	@Failure				401		{object}	stderr.ErrResponse
//	@Failure				403		{object}	stderr.ErrResponse
//	@Failure				404		{object}	stderr.ErrResponse
//	@Failure				500		{object}	stderr.ErrResponse
//	@Success				201		{object}	app.Org
//	@Router					/v1/orgs/current/user-journeys [POST]
func (s *service) CreateUserJourney(ctx *gin.Context) {
	var req CreateUserJourneyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	org, err = s.getOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, journey := range org.UserJourneys {
		if journey.Name == req.Name {
			ctx.Error(fmt.Errorf("journey with name '%s' already exists", req.Name))
			return
		}
	}

	steps := make([]app.UserJourneyStep, len(req.Steps))
	for i, stepReq := range req.Steps {
		steps[i] = app.UserJourneyStep{
			Name:     stepReq.Name,
			Title:    stepReq.Title,
			Complete: false,
		}
	}

	newJourney := app.UserJourney{
		Name:  req.Name,
		Title: req.Title,
		Steps: steps,
	}

	if org.UserJourneys == nil {
		org.UserJourneys = []app.UserJourney{}
	}
	org.UserJourneys = append(org.UserJourneys, newJourney)

	if err := s.db.WithContext(ctx).Save(org).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create user journey: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, org)
}