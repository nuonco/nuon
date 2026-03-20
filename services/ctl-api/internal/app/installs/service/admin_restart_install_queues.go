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

type RestartInstallQueuesRequest struct{}

// @ID						AdminRestartInstallQueues
// @Summary				restart all queue workflows for an install
// @Description.markdown	restart_install_queues.md
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	RestartInstallQueuesRequest	true	"Input"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id}/admin-restart-queues [POST]
func (s *service) RestartInstallQueues(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req RestartInstallQueuesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	var queues []app.Queue
	if res := s.db.WithContext(ctx).Where("owner_id = ?", install.ID).Find(&queues); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install queues: %w", res.Error))
		return
	}

	for _, queue := range queues {
		if err := s.queueClient.Restart(ctx, queue.ID); err != nil {
			ctx.Error(fmt.Errorf("unable to restart queue %s: %w", queue.ID, err))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}
