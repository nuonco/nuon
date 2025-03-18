package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type RestartAllOrgRequest struct{}

// @ID AdminRestartAll
// @Summary	restart all orgs
// @Description.markdown restart_all_orgs.md
// @Param			req		body	RestartOrgRequest	true	"Input"
// @Tags			orgs/admin
// @Security AdminEmail
// @Id AdminRestartAll
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/orgs/admin-restart-all [POST]
func (s *service) RestartAllOrgs(ctx *gin.Context) {
	var req RestartAllOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	orgs, err := s.getAllOrgs(ctx, "")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}

	for _, org := range orgs {
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type: sigs.OperationRestart,
		})
	}

	ctx.JSON(http.StatusOK, true)
}
