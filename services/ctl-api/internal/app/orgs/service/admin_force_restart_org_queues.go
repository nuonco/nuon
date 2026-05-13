package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type ForceRestartOrgQueuesRequest struct{}

// @ID						AdminForceRestartOrgQueues
// @Summary				force restart all queue workflows for an org
// @Description.markdown	force_restart_org_queues.md
// @Param					org_id	path	string							true	"org ID"
// @Param					req		body	ForceRestartOrgQueuesRequest		true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-force-restart-queues [POST]
func (s *service) ForceRestartOrgQueues(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req ForceRestartOrgQueuesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	var queues []app.Queue
	if res := s.db.WithContext(ctx).Where(app.Queue{OrgID: &org.ID}).Find(&queues); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get org queues: %w", res.Error))
		return
	}

	for _, queue := range queues {
		if err := s.queueClient.ForceRestart(ctx, queue.ID); err != nil {
			ctx.Error(fmt.Errorf("unable to force restart queue %s: %w", queue.ID, err))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}
