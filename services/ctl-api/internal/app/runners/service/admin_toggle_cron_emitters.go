package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

type AdminToggleCronEmittersRequest struct {
	OrgID         string `json:"org_id"`
	InstallID     string `json:"install_id"`
	Disable       *bool  `json:"disable" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
	StopWorkflows bool   `json:"stop_workflows"`
}

type AdminToggleCronEmittersResponse struct {
	Changed int `json:"changed"`
	Errors  int `json:"errors"`
}

// @ID						AdminToggleCronEmitters
// @Summary				Toggle cron emitters
// @Description			Disable or re-enable non-deleted cron emitters. With no org/install filters this covers every org.
// @Description			Pass org_id to scope to one org, or install_id to scope to queues owned by one install.
// @Description			disable=true stamps status disabled with reason. disable=false re-enables only emitters whose
// @Description			status_human_description matches reason. stop_workflows only applies when disabling: true stops
// @Description			emitter workflows immediately; false lets them self-exit on the next alive check / cron tick.
// @Param					req	body	AdminToggleCronEmittersRequest	true	"disable, reason, and optional filters"
// @Tags					queues/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	AdminToggleCronEmittersResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Router					/v1/runners/toggle-cron-emitters [POST]
func (s *service) AdminToggleCronEmitters(ctx *gin.Context) {
	var req AdminToggleCronEmittersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	resp, err := s.emitterClient.ToggleCronEmitters(ctx, &emitterclient.ToggleCronEmittersRequest{
		OrgID:         req.OrgID,
		InstallID:     req.InstallID,
		Disable:       *req.Disable,
		Reason:        req.Reason,
		StopWorkflows: req.StopWorkflows,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to toggle cron emitters: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, AdminToggleCronEmittersResponse{
		Changed: resp.Changed,
		Errors:  resp.Errors,
	})
}
