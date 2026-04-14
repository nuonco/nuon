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

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	var queues []app.Queue
	if res := s.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&queues); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get org queues: %w", res.Error))
		return
	}

	for _, queue := range queues {
		if err := s.queueClient.Restart(ctx, queue.ID); err != nil {
			ctx.Error(fmt.Errorf("unable to restart queue %s: %w", queue.ID, err))
			return
		}

		emitters, err := s.emitterClient.GetEmittersByQueueID(ctx, queue.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get emitters for queue %s: %w", queue.ID, err))
			return
		}

		for _, emitter := range emitters {
			if _, err := s.emitterClient.RestartEmitterWorkflow(ctx, emitter.ID); err != nil {
				ctx.Error(fmt.Errorf("unable to restart emitter %s: %w", emitter.ID, err))
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, true)
}
