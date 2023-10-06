package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

//	@BasePath	/v1/apps
//
// Delete an app
//
//	@Summary	delete an app
//	@Schemes
//	@Description	delete an app
//	@Param			app_id	path	string	true	"app ID"
//	@Tags			apps
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{boolean}	true
//	@Router			/v1/apps/{app_id} [DELETE]
func (s *service) DeleteApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	err := s.deleteApp(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, appID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteApp(ctx context.Context, appID string) error {
	currentApp := app.App{
		ID: appID,
	}

	res := s.db.WithContext(ctx).Model(&currentApp).Updates(app.App{
		Status:            "delete_queued",
		StatusDescription: "delete has been queued and waiting",
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update app: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("app not found %s: %w", appID, gorm.ErrRecordNotFound)
	}

	return nil
}
