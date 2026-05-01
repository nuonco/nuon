package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	orgrestart "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/restart"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartAllOrgRequest struct{}

// @ID						AdminRestartAll
// @Summary				restart all orgs
// @Description.markdown	restart_all_orgs.md
// @Param					req	body	RestartOrgRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Id						AdminRestartAll
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/admin-restart-all [POST]
func (s *service) RestartAllOrgs(ctx *gin.Context) {
	var req RestartAllOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	orgs, err := s.getAllOrgs(ctx, "")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get orgs: %w", err))
		return
	}

	for _, org := range orgs {
		useQueues, err := s.useOrgQueues(ctx, org.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("checking features for org %s: %w", org.ID, err))
			return
		}
		if useQueues {
			queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
			if err != nil {
				ctx.Error(fmt.Errorf("unable to get org signals queue for org %s: %w", org.ID, err))
				return
			}
			if err := s.enqueueOrgSignal(ctx, queueID, &orgrestart.Signal{OrgID: org.ID}, org.ID); err != nil {
				ctx.Error(fmt.Errorf("enqueue signal for org %s: %w", org.ID, err))
				return
			}
		} else {
			s.evClient.Send(ctx, org.ID, &sigs.Signal{
				Type: sigs.OperationRestart,
			})
		}
	}

	ctx.JSON(http.StatusOK, true)
}
