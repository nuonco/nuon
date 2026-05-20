package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgforcerestartqueues "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/force_restart_queues"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
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
// @Success				200	{object}	map[string]string
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

	queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org signals queue: %w", err))
		return
	}

	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    &orgforcerestartqueues.Signal{OrgID: org.ID},
		OwnerID:   org.ID,
		OwnerType: plugins.TableName(s.db, app.Org{}),
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue force restart signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"queue_signal_id": resp.ID,
		"queue_id":        queueID,
	})
}
