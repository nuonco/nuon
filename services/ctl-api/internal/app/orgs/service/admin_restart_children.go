package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type RestartOrgChildrenRequest struct{}

//	@ID						AdminRestartOrgChildren
//	@Summary				restart an org and all it's children event loops
//	@Description.markdown	restart_org_children.md
//	@Param					org_id	path	string						true	"org ID"
//	@Param					req		body	RestartOrgChildrenRequest	true	"Input"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{boolean}	true
//	@Router					/v1/orgs/{org_id}/admin-restart-children [POST]
func (s *service) RestartOrgChildren(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req RestartOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}
	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationRestartChildren,
	})

	ctx.JSON(http.StatusOK, true)
}
