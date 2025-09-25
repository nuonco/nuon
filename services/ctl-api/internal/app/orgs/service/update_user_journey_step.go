package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type UpdateUserJourneyStepRequest struct {
	Complete bool `json:"complete" binding:""`
}

//	@ID						UpdateOrgUserJourneyStep
//	@Summary				Update user journey step completion status
//	@Description			Mark a user journey step as complete or incomplete
//	@Tags					orgs
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Param					journey_name	path		string								true	"Journey name"
//	@Param					step_name		path		string								true	"Step name"
//	@Param					body			body		UpdateUserJourneyStepRequest		true	"Update step request"
//	@Failure				400				{object}	stderr.ErrResponse
//	@Failure				401				{object}	stderr.ErrResponse
//	@Failure				403				{object}	stderr.ErrResponse
//	@Failure				404				{object}	stderr.ErrResponse
//	@Failure				500				{object}	stderr.ErrResponse
//	@Success				200				{object}	app.Org
//	@Router					/v1/orgs/current/user-journeys/{journey_name}/steps/{step_name} [PATCH]
func (s *service) UpdateUserJourneyStep(ctx *gin.Context) {
	var req UpdateUserJourneyStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	journeyName := ctx.Param("journey_name")
	stepName := ctx.Param("step_name")

	org, err = s.getOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	found := false
	for i, journey := range org.UserJourneys {
		if journey.Name == journeyName {
			for j, step := range journey.Steps {
				if step.Name == stepName {
					org.UserJourneys[i].Steps[j].Complete = req.Complete
					found = true
					break
				}
			}
			break
		}
	}

	if !found {
		ctx.Error(fmt.Errorf("journey '%s' or step '%s' not found", journeyName, stepName))
		return
	}

	if err := s.db.WithContext(ctx).Save(org).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update user journey step: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, org)
}