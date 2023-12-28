package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReprovisionOrgRequest struct{}

// @ID AdminReprovisionOrg
// @Summary reprovision an org
// @Description.markdown reprovision_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	ReprovisionOrgRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-reprovision [POST]
func (s *service) AdminReprovisionOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Reprovision(ctx, orgID)
	ctx.JSON(http.StatusOK, true)
}
