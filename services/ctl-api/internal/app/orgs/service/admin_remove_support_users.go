package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						AdminRemoveSupportUsers
// @BasePath				/v1/orgs
// @Summary				Remove nuon users as support members
// @Description.markdown	admin_remove_support_users.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-remove-support-users [POST]
func (s *service) RemoveSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.RemoveSupportUsersFromOrg(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "ok",
	})
}
