package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/forgotten"
)

type AdminForgetOrgInstallsRequest struct{}

// @ID						AdminForgetOrgInstalls
// @Summary				forget all installs for an org
// @Description.markdown	forget_org_installs.md
// @Param					org_id	path	string							true	"org ID"
// @Param					req		body	AdminForgetOrgInstallsRequest	true	"Input"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-forget-installs [POST]
func (s *service) ForgetOrgInstalls(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	installs, err := s.getOrgInstalls(ctx, orgID, "", nil, "")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org installs: %w", err))
		return
	}

	for _, install := range installs {
		err := s.forgetInstall(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, queueID, &forgotten.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}
