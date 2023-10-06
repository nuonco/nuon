package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth"
)

//	@BasePath	/v1/general

// Get current user
//
//	@Summary	Get current user
//	@Schemes
//	@Description	get current user
//	@Tags			general
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	app.UserToken
//	@Router			/v1/general/current-user [GET]
func (s *service) GetCurrentUser(ctx *gin.Context) {
	userToken, err := auth.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, userToken)
}
