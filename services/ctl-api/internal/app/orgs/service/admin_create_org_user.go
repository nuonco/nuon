package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminCreateOrgUserRequest struct {
	Email string `json:"email"`
}

// @ID AdminAddOrgUser
// @BasePath	/v1/orgs
// @Summary	Add a user to an org
// @Description.markdown create_org_user.md
// @Param			org_id	path	string	true	"org ID to add user too"
// @Tags			orgs/admin
// @Param			req	body	AdminCreateOrgUserRequest	true	"Input"
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-add-user [POST]
func (s *service) CreateOrgUser(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	var req AdminCreateOrgUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	err = s.createUserByEmail(ctx, org, req.Email)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) createUserByEmail(ctx context.Context, org *app.Org, email string) error {
	var acct app.Account
	err := s.db.WithContext(ctx).First(&acct, "email = ?", email).Error
	if err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}

	return s.authzClient.AddAccountRole(ctx, app.RoleTypeOrgAdmin, org.ID, acct.ID)
}
