package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type OrgAddLogoRequest struct {
	LogoURL string `validate:"required"`
}

//	@ID						AdminAddOrgLogo
//	@Summary				add a custom logo for an org
//	@Description.markdown	admin_add_org_logo.md
//	@Param					org_id	path	string	true	"org ID or name to update"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	OrgAddLogoRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/{org_id}/admin-add-logo [POST]
func (s *service) AdminAddLogo(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req OrgAddLogoRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.addOrgLogo(ctx, org.ID, req.LogoURL); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, true)
}

func (s *service) addOrgLogo(ctx context.Context, orgID string, logoURL string) error {
	org := app.Org{
		ID: orgID,
	}
	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		LogoURL: logoURL,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
