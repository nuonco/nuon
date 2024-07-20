package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	authcontext "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type CreateOrgUserRequest struct {
	UserID string `json:"user_id"`
}

// @ID AddUser
// @Summary	Add a user to the current org
// @Description.markdown	create_org_user.md
// @Param			req	body	CreateOrgUserRequest	true	"Input"
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.Account
// @Router			/v1/orgs/current/user [POST]
func (s *service) CreateUser(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := authcontext.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, org.ID, acct.ID); err != nil {
		ctx.Error(err)
		return
	}

	var req CreateOrgUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, acct)
}
