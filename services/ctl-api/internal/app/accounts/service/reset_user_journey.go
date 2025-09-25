package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						ResetUserJourney
// @Summary				Reset user journey steps
// @Description			Reset all steps in a specified user journey by setting their completion status to false
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					journey_name	path		string								true	"Journey name"
// @Failure				400				{object}	stderr.ErrResponse
// @Failure				401				{object}	stderr.ErrResponse
// @Failure				403				{object}	stderr.ErrResponse
// @Failure				404				{object}	stderr.ErrResponse
// @Failure				500				{object}	stderr.ErrResponse
// @Success				200				{object}	app.Account
// @Router					/v1/account/user-journeys/{journey_name}/reset [POST]
func (s *service) ResetUserJourney(ctx *gin.Context) {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	journeyName := ctx.Param("journey_name")

	// Delegate business logic to private method
	updatedAccount, err := s.resetUserJourney(ctx, account.ID, journeyName)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, updatedAccount)
}

func (s *service) resetUserJourney(ctx *gin.Context, accountID, journeyName string) (*app.Account, error) {
	// Get full account with user journeys
	fullAccount, err := s.getAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Find the specified journey and reset all steps
	found := false
	updated := false
	for i, journey := range fullAccount.UserJourneys {
		if journey.Name == journeyName {
			found = true
			for j, step := range journey.Steps {
				if step.Complete {
					fullAccount.UserJourneys[i].Steps[j].Complete = false
					updated = true
				}
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("journey '%s' not found", journeyName)
	}

	// Save if we found and updated any steps
	if updated {
		if err := s.db.WithContext(ctx).Select("user_journeys").Save(fullAccount).Error; err != nil {
			return nil, fmt.Errorf("unable to reset user journey: %w", err)
		}
	}

	return fullAccount, nil
}
