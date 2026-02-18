package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartOrgRequest struct{}

// @ID						AdminRestartOrg
// @Summary				restart an orgs event loop
// @Description.markdown	restart_org.md
// @Param					org_id	path	string				true	"org ID"
// @Param					req		body	RestartOrgRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Id						AdminRestartOrg
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-restart [POST]
func (s *service) RestartOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req RestartOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}
