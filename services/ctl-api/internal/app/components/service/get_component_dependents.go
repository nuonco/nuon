package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type ComponentChildren struct {
	Children []app.Component `json:"children"`
}

// @ID						GetComponentDependents
// @Summary					get a component's children
// @Description.markdown	get_component_dependents.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	ComponentChildren
// @Router					/v1/components/{component_id}/dependents [get]
func (s *service) GetComponentDependents(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	component, err := s.getComponent(ctx, componentID)
	if component == nil {
		ctx.Error(fmt.Errorf("component %s not found", componentID))
		return
	}

	appID := component.AppID
	components, err := s.getAppComponents(ctx, appID, "", nil)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get components for app: %s", appID))
		return
	}

	cdo := s.appsHelpers.GetComponentsDependents(componentID, components)

	ctx.JSON(http.StatusOK, ComponentChildren{
		Children: cdo,
	})
}
