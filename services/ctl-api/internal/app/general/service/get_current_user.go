package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authcontext "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetCurrentUser
// @Summary	Get current user
// @Description.markdown	get_current_user.md
// @Tags			general
// @Accept			json
// @Produce		json
// @Security APIKey
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.Account
// @Router			/v1/general/current-user [GET]
func (s *service) GetCurrentUser(ctx *gin.Context) {
	acct, err := authcontext.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, acct)
}
