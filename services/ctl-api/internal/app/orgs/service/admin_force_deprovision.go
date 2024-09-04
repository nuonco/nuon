package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type AdminForceDeprovisionOrgRequest struct{}

// @ID AdminForceDeprovisionOrg
// @Summary force deprovision an org, without waiting for dependencies
// @Description.markdown force_deprovision_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	AdminForceDeprovisionOrgRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-force-deprovision [POST]
func (s *service) AdminForceDeprovisionOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationForceDeprovision,
	})

	ctx.JSON(http.StatusCreated, true)
}
