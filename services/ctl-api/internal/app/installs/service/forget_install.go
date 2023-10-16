package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminForgetInstallRequest struct{}

//	@BasePath	/v1/installs
//
// Forget an install that has been deleted outside of nuon.
// This should only be used in cases where an install was broken in an unordinary way and needs to be manually deleted
// so the parent resources can be deleted.
//
//	@Summary	forget an install
//	@Schemes
//	@Description	forget an install
//	@Param			install_id	path	string						true	"install ID"
//	@Param			req			body	AdminForgetInstallRequest	true	"Input"
//	@Tags			installs/admin
//	@Accept			json
//	@Produce		json
//	@Failure		400	{object}	stderr.ErrResponse
//	@Failure		404	{object}	stderr.ErrResponse
//	@Failure		500	{object}	stderr.ErrResponse
//	@Success		200	{boolean}	true
//	@Router			/v1/installs/{install_id}/admin-forget [POST]
func (s *service) ForgetInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	err := s.forgetInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Forgotten(ctx, installID)
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
		return fmt.Errorf("install not found %s", installID)
	}
	return nil
}
