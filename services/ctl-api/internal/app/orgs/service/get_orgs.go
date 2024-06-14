package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authcontext "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth/context"
)

// @ID GetOrgs
// @Summary	Return current user's orgs
// @Description.markdown get_orgs.md
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.Org
// @Router			/v1/orgs [GET]
func (s *service) GetCurrentUserOrgs(ctx *gin.Context) {
	account, err := authcontext.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, account.Orgs)
}
