package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/services/config"
	orgforcesandboxmode "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/force_sandbox_mode"
	orgrestart "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/restart"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org signals queue: %w", err))
		return
	}
	if err := s.enqueueOrgSignal(ctx, queueID, &orgrestart.Signal{OrgID: org.ID}, org.ID); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	if err := s.enqueueOrgSignal(ctx, queueID, &orgforcesandboxmode.Signal{OrgID: org.ID}, org.ID); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, true)
}
