package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/registryprune"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type AdminRegistryPruneRequest struct {
	DryRun bool `json:"dry_run"`
}

// @ID						AdminRegistryPrune
// @Summary				enqueue install registry prune
// @Description			Enqueues registry-prune for this install. With dry_run it reports what would be deleted, even when registry retention is off.
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	AdminRegistryPruneRequest	true	"Input"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	map[string]string
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/installs/{install_id}/admin-registry-prune [POST]
func (s *service) AdminRegistryPrune(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AdminRegistryPruneRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	install, err := s.getInstallCrossOrg(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	queueID, err := s.getInstallQueueID(ctx, install.ID, helpers.InstallDriftCronSignalsQueueName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install drift cron queue: %w", err))
		return
	}

	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    &registryprune.Signal{InstallID: install.ID, DryRun: req.DryRun},
		OwnerID:   install.ID,
		OwnerType: plugins.TableName(s.db, app.Install{}),
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue registry prune: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"queue_signal_id": resp.ID,
		"queue_id":        queueID,
	})
}
