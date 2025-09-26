package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						CompleteUserJourney
// @Summary				Complete all steps in a specific user journey
// @Description			Mark all remaining steps in the specified user journey as complete
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Param					journey_name	path	string	true	"Journey name to complete"
// @Security				APIKey
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/account/user-journeys/{journey_name}/complete [POST]
func (s *service) CompleteUserJourney(ctx *gin.Context) {
	journeyName := ctx.Param("journey_name")
	if journeyName == "" {
		ctx.Error(fmt.Errorf("journey_name parameter is required"))
		return
	}
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Delegate business logic to private method
	updatedAccount, err := s.completeUserJourney(ctx, account.ID, journeyName)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, updatedAccount)
}

func (s *service) completeUserJourney(ctx *gin.Context, accountID string, journeyName string) (*app.Account, error) {
	// Get full account with user journeys
	fullAccount, err := s.getAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Find the specified journey and complete all steps
	found := false
	updated := false

	for i, journey := range fullAccount.UserJourneys {
		if journey.Name == journeyName {
			found = true
			for j, step := range journey.Steps {
				if !step.Complete {
					fullAccount.UserJourneys[i].Steps[j].Complete = true
					updated = true
				}
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("journey '%s' not found for account", journeyName)
	}

	// Save if we found and updated any steps
	if updated {
		if err := s.db.WithContext(ctx).Select("user_journeys").Save(fullAccount).Error; err != nil {
			return nil, fmt.Errorf("unable to complete user journey '%s': %w", journeyName, err)
		}
	}

	return fullAccount, nil
}
