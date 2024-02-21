package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var defaultSupportUsers = []string{
	// Dre Smith
	"google-oauth2|113884954942864770921",
	//Jon Morehouse
	"google-oauth2|114670241124324496631",
	//Jordan Acosta
	"google-oauth2|106750268626571499868",
	//Nat Hamilton
	"google-oauth2|107796233904597398271",
	// Pavi Sandhu
	"google-oauth2|117375967099708763726",
}

// @ID AdminCreateSupportUsers
// @BasePath	/v1/orgs
// @Summary	Add nuon users as support members
// @Description.markdown create_org_support_users.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-support-users [POST]
func (s *service) CreateSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cctx := context.WithValue(ctx, "user_id", org.CreatedByID)
	for _, userID := range defaultSupportUsers {
		if _, err := s.createUser(cctx, orgID, userID); err != nil {
			ctx.Error(fmt.Errorf("unable to add users to org: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
