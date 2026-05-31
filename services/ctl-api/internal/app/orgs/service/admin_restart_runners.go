package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	orgrestartrunners "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/restart_runners"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartOrgRunnersRequest struct{}

// @ID						AdminRestartOrgRunners
// @Summary				restart all runners in an org
// @Description.markdown	restart_org_runners.md
// @Param					org_id	path	string						true	"org ID"
// @Param					req		body	RestartOrgRunnersRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-restart-runners [POST]
func (s *service) AdminRestartRunners(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

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
	if err := s.enqueueOrgSignal(ctx, queueID, &orgrestartrunners.Signal{OrgID: org.ID}, org.ID); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, true)
}
