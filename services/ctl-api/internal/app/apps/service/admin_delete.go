package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	installsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminDeleteAppRequest struct{}

// @ID AdminDeleteApp
// @Summary	delete an app
// @Description.markdown delete_app.md
// @Tags			apps/admin
// @Accept			json
// @Param			req		body	AdminDeleteAppRequest	true	"Input"
// @Param			app_id	path	string					true	"app id"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/apps/{app_id}/admin-delete [POST]
func (s *service) AdminDeleteApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	currentApp, err := s.getAppAndDependencies(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	for _, install := range currentApp.Installs {
		s.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationDelete,
		})
		s.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationForgotten,
		})
	}

	for _, comp := range currentApp.Components {
		s.evClient.Send(ctx, comp.ID, &componentsignals.Signal{
			Type: componentsignals.OperationDelete,
		})
	}
	s.evClient.Send(ctx, appID, &signals.Signal{
		Type: signals.OperationDeleted,
	})
	s.evClient.Send(ctx, appID, &signals.Signal{
		Type: signals.OperationDeprovision,
	})

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getAppAndDependencies(ctx context.Context, appID string) (*app.App, error) {
	currentApp := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Installs").
		Preload("Components").
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app %s: %w", appID, res.Error)
	}

	return &currentApp, nil
}
