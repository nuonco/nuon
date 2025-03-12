package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID GetAppComponents
// @Summary	get all components for an app
// @Description.markdown	get_app_components.md
// @Param			app_id	path	string	true	"app ID"
// @Param   offset query int	 false	"offset of results to return"	Default(0)
// @Param   limit  query int	 false	"limit of results to return"	     Default(10)
// @Param   x-nuon-pagination-enabled header bool false "Enable pagination"
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
// @Success		200				{array}		app.Component
// @Router			/v1/apps/{app_id}/components [GET]
func (s *service) GetAppComponents(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	component, err := s.getAppComponents(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app components: %w", err))
		return
	}

	reorderedCmp, err := s.appsHelpers.OrderComponentsByDep(ctx, component)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to order components by dependency: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, reorderedCmp)
}

func (s *service) getAppComponents(ctx *gin.Context, appID string) ([]app.Component, error) {
	currentApp := &app.App{}
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithPagination).
		Preload("Components").
		Preload("Components.ComponentConfigs").
		Preload("Components.Dependencies").
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	cmps, err := db.HandlePaginatedResponse(ctx, currentApp.Components)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	currentApp.Components = cmps

	return currentApp.Components, nil
}
