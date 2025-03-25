package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type RestartOrgRunnersRequest struct{}

//	@ID						AdminRestartOrgRunners
//	@Summary				restart all runners in an org
//	@Description.markdown	restart_org_runners.md
//	@Param					org_id	path	string						true	"org ID"
//	@Param					req		body	RestartOrgRunnersRequest	true	"Input"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{boolean}	true
//	@Router					/v1/orgs/{org_id}/admin-restart-runners [POST]
func (s *service) AdminRestartRunners(ctx *gin.Context) {
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
		Type: sigs.OperationRestartRunners,
	})
	ctx.JSON(http.StatusOK, true)
}
