package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/services/config"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type OrgForceSandboxModeRequest struct{}

// @ID						AdminOrgForceSandboxMode
// @Summary				force an org into sandbox mode
// @Description.markdown org_force_sandbox_mode.md
// @Param					org_id	path	string				true	"org ID"
// @Param					req		body	OrgForceSandboxModeRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-force-sandbox-mode [POST]
func (s *service) AdminForceSandboxMode(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	if s.cfg.Env != config.Development {
		ctx.Error(fmt.Errorf("this endpoint is only supported in local development"))
		return
	}

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
		Type: sigs.OperationRestart,
	})
	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationForceSandboxMode,
	})
	ctx.JSON(http.StatusOK, true)
}
