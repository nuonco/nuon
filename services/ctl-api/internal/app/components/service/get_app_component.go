package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetAppComponent
// @Summary	get a components for a specific app
// @Description.markdown	get_app_component.md
// @Param			app_id	path	string	true	"app ID"
// @Param			component_name_or_id	path	string	true	"name or ID"
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}		app.Component
// @Router			/v1/apps/{app_id}/component/{component_name_or_id} [GET]
func (s *service) GetAppComponent(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	idOrName := ctx.Param("component_name_or_id")

	component, err := s.getAppComponent(ctx, appID, idOrName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app component: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) getAppComponent(ctx context.Context, appID, componentNameOrID string) (*app.Component, error) {
	component := app.Component{}
	res := s.db.WithContext(ctx).
		Where("app_id = ?", appID).
		Where(
			s.db.Where("id = ?", componentNameOrID).
				Or("name = ?", componentNameOrID),
		).
		First(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	comp, err := s.helpers.GetComponent(ctx, component.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return comp, nil
}
