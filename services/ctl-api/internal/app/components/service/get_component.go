package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/components
// Get a component
//	@Summary	get a component
//	@Schemes
//	@Description	get a component
//	@Param			component_id	path	string	true	"component ID"
//	@Tags			components
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	app.Component
//	@Router			/v1/components/{component_id} [get]
func (s *service) GetComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  component %s: %w", componentID, err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) getComponent(ctx context.Context, componentID string) (*app.Component, error) {
	component := app.Component{}
	res := s.db.WithContext(ctx).
		Where("id = ?", componentID).
		Or("name = ?", componentID).
		Preload("ComponentConfigs").
		First(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}
	component.ConfigVersions = len(component.ComponentConfigs)

	return &component, nil
}

func (s *service) getComponentWithParents(ctx context.Context, cmpID string) (*app.Component, error) {
	parentCmp := app.Component{}
	res := s.db.WithContext(ctx).Preload("App").Preload("App.Org").Preload("App.Org.VCSConnections").First(&parentCmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &parentCmp, nil
}
