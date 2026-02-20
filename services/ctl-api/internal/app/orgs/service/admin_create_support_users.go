package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						AdminCreateSupportUsers
// @BasePath				/v1/orgs
// @Summary				Add nuon users as support members
// @Description.markdown	create_org_support_users.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-support-users [POST]
func (s *service) CreateSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	results, err := s.helpers.AddSupportUsersToOrg(ctx, org)
	if err != nil {
		ctx.Error(err)
		return
	}

	successCount := 0
	alreadyExistsCount := 0
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
		} else if result.AlreadyExists {
			alreadyExistsCount++
		} else if result.Success {
			successCount++
		}
	}

	ctx.JSON(http.StatusCreated, map[string]interface{}{
		"status":              "ok",
		"success":             successCount,
		"already_exists":      alreadyExistsCount,
		"errors":              errorCount,
		"total_support_users": len(results),
	})
}
