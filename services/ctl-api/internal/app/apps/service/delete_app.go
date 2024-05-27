package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
)

// @ID DeleteApp
// @Summary	delete an app
// @Description.markdown	delete_app.md
// @Param			app_id	path	string	true	"app ID"
// @Tags			apps
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{boolean}	true
// @Router			/v1/apps/{app_id} [DELETE]
func (s *service) DeleteApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	err := s.deleteApp(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, appID, &signals.Signal{
		Type: signals.OperationDeleted,
	})
	s.evClient.Send(ctx, appID, &signals.Signal{
		Type: signals.OperationDeprovision,
	})
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
