package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminForgetOrgRequest struct{}

// @ID						AdminForgetOrg
// @Summary				forget an org and everything in it
// @Description.markdown	admin_forget_org.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminForgetOrgRequest	true	"Input"
// @Produce				json
// @Success				202	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-forget [POST]
func (s *service) AdminForgetOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminForgetOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Soft delete roles (and their join-table entries) so the Account AfterQuery
	// hook no longer tries to dereference the now-deleted org.
	if err := s.db.WithContext(ctx).Where("org_id = ?", org.ID).Delete(&app.Role{}).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to forget org roles: %w", err))
		return
	}

	res := s.db.WithContext(ctx).Delete(&app.Org{ID: org.ID})
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to forget org: %w", res.Error))
		return
	}
	if res.RowsAffected < 1 {
		ctx.Error(fmt.Errorf("org not found %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusAccepted, true)
}
