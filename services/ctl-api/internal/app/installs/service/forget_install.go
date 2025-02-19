package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type ForgetInstallRequest struct{}

// @ID ForgetInstall
// @Summary	forget an install
// @Description.markdown forget_install.md
// @Param		install_id	path	string						true	"install ID"
// @Param		req			body	ForgetInstallRequest	true	"Input"
// @Tags		installs
// @Security APIKey
// @Security OrgID
// @Accept			json
// @Produce		json
// @Failure		400	{object}	stderr.ErrResponse
// @Failure		404	{object}	stderr.ErrResponse
// @Failure		500	{object}	stderr.ErrResponse
// @Success		200	{boolean}	true
// @Router			/v1/installs/{install_id}/forget [POST]
func (s *service) ForgetInstall(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	err = s.forgetInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationForgotten,
	})
	ctx.JSON(http.StatusOK, true)
}
