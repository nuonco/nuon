package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type RenameOrgRequest struct {
	Name string `validate:"required"`
}

// @ID AdminRenameOrg
// @Summary rename an org
// @Description.markdown rename_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	RenameOrgRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-rename [POST]
func (s *service) AdminRenameOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req RenameOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.renameOrg(ctx, orgID, req.Name); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) renameOrg(ctx context.Context, orgID string, name string) error {
	org := app.Org{
		ID: orgID,
	}
	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		Name: name,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
