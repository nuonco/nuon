package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type RestartOrgRequest struct{}

// @ID AdminRestartOrg
// @Summary	restart an orgs event loop
// @Description.markdown restart_org.md
// @Param			org_id	path	string				true	"org ID"
// @Param			req		body	RestartOrgRequest	true	"Input"
// @Tags			orgs/admin
// @Id AdminRestartOrg
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/orgs/{org_id}/admin-restart [POST]
func (s *service) RestartOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req RestartOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}
