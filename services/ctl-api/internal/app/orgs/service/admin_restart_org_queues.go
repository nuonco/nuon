package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	orgrestartqueues "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/restart_queues"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartOrgQueuesRequest struct{}

// @ID						AdminRestartOrgQueues
// @Summary				restart all queue workflows for an org
// @Description.markdown	restart_org_queues.md
// @Param					org_id	path	string					true	"org ID"
// @Param					req		body	RestartOrgQueuesRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-restart-queues [POST]
func (s *service) RestartOrgQueues(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req RestartOrgQueuesRequest
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
	if err := s.enqueueOrgSignal(ctx, queueID, &orgrestartqueues.Signal{OrgID: org.ID}); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
