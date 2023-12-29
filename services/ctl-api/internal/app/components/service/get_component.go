package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @ID GetComponent
// @Summary	get a component
// @Description.markdown	get_component.md
// @Param			component_id	path	string	true	"component ID"
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
// @Success		200				{object}	app.Component
// @Router			/v1/components/{component_id} [get]
func (s *service) GetComponent(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")

	component, err := s.findComponent(ctx, org.ID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component %s: %w", componentID, err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) findComponent(ctx context.Context, orgID, componentID string) (*app.Component, error) {
	component := app.Component{}
	res := s.db.WithContext(ctx).
		Where("id = ?", componentID).
		Or("name = ? AND org_id = ?", componentID, orgID).
		Preload("ComponentConfigs").
		Preload("Dependencies").
		Preload("App").
		Preload("App.Org").
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
