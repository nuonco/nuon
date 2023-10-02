package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/installs
//
// Delete an install
//
//	@Summary	delete an install
//	@Schemes
//	@Description	delete an install
//	@Param			install_id	path	string	true	"install ID"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{boolean}	true
//	@Router			/v1/installs/{install_id} [DELETE]
func (s *service) DeleteInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	err := s.deleteInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, installID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteInstall(ctx context.Context, installID string) error {
	install := app.Install{
		ID: installID,
	}

	res := s.db.WithContext(ctx).Model(&install).Updates(app.Install{
		Status:            "delete_queued",
		StatusDescription: "delete has been queued and waiting",
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("install not found %s", installID)
	}

	return nil
}
