package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminForgetInstallRequest struct{}

// @ID AdminForgetInstall
// @Summary	forget an install
// @Description.markdown forget_install.md
// @Param		install_id	path	string						true	"install ID"
// @Param			req			body	AdminForgetInstallRequest	true	"Input"
// @Tags			installs/admin
// @Accept			json
// @Produce		json
// @Failure		400	{object}	stderr.ErrResponse
// @Failure		404	{object}	stderr.ErrResponse
// @Failure		500	{object}	stderr.ErrResponse
// @Success		200	{boolean}	true
// @Router			/v1/installs/{install_id}/admin-forget [POST]
func (s *service) ForgetInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
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

func (s *service) forgetInstall(ctx context.Context, installID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Install{
		ID: installID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete install: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("install not found %s %s", installID, gorm.ErrRecordNotFound)
	}
	return nil
}
