package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type AdminDeprovisionOrgRequest struct {
	Force bool `json:"force"`
}

//	@ID						AdminDeprovisionOrg
//	@Summary				deprovision an org, but keep it in the database
//	@Description.markdown	deprovision_org.md
//	@Param					org_id	path	string	true	"org ID for your current org"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	AdminDeprovisionOrgRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/{org_id}/admin-deprovision [POST]
func (s *service) AdminDeprovisionOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")
	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminDeprovisionOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	sigTyp := sigs.OperationDeprovision
	if req.Force {
		sigTyp = sigs.OperationForceDeprovision
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigTyp,
	})

	ctx.JSON(http.StatusOK, true)
}
