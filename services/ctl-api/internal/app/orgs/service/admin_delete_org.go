package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type AdminDeleteOrgRequest struct {
	Force bool `json:"force"`
}

//	@ID						AdminDeleteOrg
//	@Summary				delete an org and everything in it
//	@Description.markdown	delete_org.md
//	@Param					org_id	path	string	true	"org ID for your current org"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	AdminDeleteOrgRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/{org_id}/admin-delete [POST]
func (s *service) AdminDeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminDeleteOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if org.OrgType == app.OrgTypeIntegration {
		err := s.helpers.HardDelete(ctx, org.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, true)
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type:        sigs.OperationDelete,
		ForceDelete: req.Force,
	})

	ctx.JSON(http.StatusOK, true)
}
