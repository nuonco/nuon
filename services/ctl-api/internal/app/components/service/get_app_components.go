package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/apps/components
//
// Get an app's components
//
//	@Summary	get all components for an app
//	@Schemes
//	@Description	get all components for an org
//	@Param			app_id	path	string	true	"app ID"
//	@Tags			components
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{array}		app.Component
//	@Router			/v1/apps/{app_id}/components [GET]
func (s *service) GetAppComponents(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	component, err := s.getAppComponents(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) getAppComponents(ctx context.Context, appID string) ([]app.Component, error) {
	currentApp := &app.App{}
	res := s.db.WithContext(ctx).
		Preload("Components").
		Preload("Components.ComponentConfigs").
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	comps := currentApp.Components
	for idx, comp := range comps {
		comps[idx].ConfigVersions = len(comp.ComponentConfigs)
	}

	return comps, nil
}
