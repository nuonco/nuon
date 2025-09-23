package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
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
	appCfg, err := s.appsHelpers.GetAppLatestConfig(ctx, appID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get app latest config"))
		return
	}

	if !generics.SliceContains(component.ID, appCfg.ComponentIDs) {
		ctx.Error(errors.Wrap(err, "component does not belong to a current app config"))
		return
	}

	cdo, err := s.appsHelpers.GetComponentDependents(ctx, appCfg.ID, componentID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get component dependents"))
		return
	}

	comps, err := s.appsHelpers.GetComponents(ctx, cdo)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get components"))
		return
	}

	ctx.JSON(http.StatusOK, ComponentChildren{
		Children: comps,
	})
}
