package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminDeprovisionOrgRequest struct{}

// @ID AdminDeprovisionOrg
// @Summary deprovision an install, but keep it in the database
// @Description.markdown deprovision_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	AdminDeprovisionOrgRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-deprovision [POST]
func (s *service) AdminDeprovisionOrg(ctx *gin.Context) {
	installID := ctx.Param("org_id")
	s.hooks.Deprovisioned(ctx, installID)
	ctx.JSON(http.StatusOK, true)
}
