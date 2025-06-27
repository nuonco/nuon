package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	for _, user := range defaultSupportUsers {
		if err := s.removeSupportUser(ctx, user[0], org.ID); err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "ok",
	})
}

func (s *service) removeSupportUser(ctx context.Context, email, orgID string) error {
	acct, err := s.acctClient.FindAccount(ctx, email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return nil
	}

	if err := s.authzClient.RemoveAccountOrgRoles(ctx, orgID, acct.ID); err != nil {
		return err
	}

	return nil
}
