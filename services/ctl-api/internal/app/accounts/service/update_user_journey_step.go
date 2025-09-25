package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type UpdateUserJourneyStepRequest struct {
	Complete bool `json:"complete" binding:""`
}

// @ID						UpdateUserJourneyStep
// @Summary				Update user journey step completion status
// @Description			Mark a user journey step as complete or incomplete
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Param					journey_name	path		string								true	"Journey name"
// @Param					step_name		path		string								true	"Step name"
// @Param					body			body		UpdateUserJourneyStepRequest		true	"Update step request"
// @Failure				400				{object}	stderr.ErrResponse
// @Failure				401				{object}	stderr.ErrResponse
// @Failure				403				{object}	stderr.ErrResponse
// @Failure				404				{object}	stderr.ErrResponse
// @Failure				500				{object}	stderr.ErrResponse
// @Success				200				{object}	app.Account
// @Router					/v1/account/user-journeys/{journey_name}/steps/{step_name} [PATCH]
func (s *service) UpdateUserJourneyStep(ctx *gin.Context) {
	var req UpdateUserJourneyStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	journeyName := ctx.Param("journey_name")
	stepName := ctx.Param("step_name")

	// Delegate business logic to private method
	updatedAccount, err := s.updateUserJourneyStep(ctx, account.ID, journeyName, stepName, req.Complete)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, updatedAccount)
}

func (s *service) updateUserJourneyStep(ctx *gin.Context, accountID, journeyName, stepName string, complete bool) (*app.Account, error) {
	account, err := s.getAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Find and update the specific journey step
	found := false
	for i, journey := range account.UserJourneys {
		if journey.Name == journeyName {
			for j, step := range journey.Steps {
				if step.Name == stepName {
					account.UserJourneys[i].Steps[j].Complete = complete
					found = true
					break
				}
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("journey '%s' or step '%s' not found", journeyName, stepName)
	}

	// Save to database
	if err := s.db.WithContext(ctx).Select("user_journeys").Save(account).Error; err != nil {
		return nil, fmt.Errorf("unable to update user journey step: %w", err)
	}

	return account, nil
}
