package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminForgetOrgRequest struct{}

// @ID AdminForgetOrg
// @Summary forget an org and everything in it
// @Description.markdown forget_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Security AdminEmail
// @Accept			json
// @Param			req	body	AdminForgetOrgRequest	true	"Input"
// @Produce		json
// @Success		202	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-forget [POST]
func (s *service) AdminForgetOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminForgetOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.HardDelete(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, true)
}
