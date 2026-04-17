package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						GetOnboardingExampleApps
// @Summary				Get example apps catalog
// @Description			Returns the list of available example applications for onboarding
// @Tags					onboarding
// @Produce				json
// @Security				APIKey
// @Success				200	{array}		ExampleApp
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/example-apps [GET]
func (s *service) GetExampleApps(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, s.catalog.Get(ctx.Request.Context()))
}
