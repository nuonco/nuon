package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/tokencleanup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						AdminCleanupInstallTokens
// @Summary				enqueue install token cleanup
// @Description			Enqueues install-token-cleanup on this install's install-signals queue. The signal deletes terminal job accounts and revokes credentials for closed or deleted log streams.
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	map[string]string
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/installs/{install_id}/admin-cleanup-tokens [POST]
func (s *service) AdminCleanupInstallTokens(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstallCrossOrg(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install-signals queue: %w", err))
		return
	}

	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    &tokencleanup.Signal{InstallID: install.ID},
		OwnerID:   install.ID,
		OwnerType: plugins.TableName(s.db, app.Install{}),
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue install token cleanup: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"queue_signal_id": resp.ID,
		"queue_id":        queueID,
	})
}
