package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type ComponentChildren struct {
	Children []app.Component `json:"children"`
}

// @ID						GetAppComponentDependents
// @Summary					get a component's children
// @Description.markdown	get_component_dependents.md
// @Param					app_id			path	string	true	"app ID"
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
// @Router					/v1/apps/{app_id}/components/{component_id}/dependents [get]
func (s *service) GetAppComponentDependents(ctx *gin.Context) (any, error) {
	return s.GetComponentDependents(ctx)
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
func (s *service) GetComponentDependents(ctx *gin.Context) (any, error) {
	components, err := s.readComponentDependents(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	return ComponentChildren{Children: components}, nil
}

func (s *service) readComponentDependents(ctx context.Context) ([]app.Component, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := access.Target("component_id")
	if err != nil {
		return nil, err
	}
	query, err := access.Query()
	if err != nil {
		return nil, err
	}
	var component app.Component
	if err := access.DB().WithContext(ctx).Where(&app.Component{ID: target.ID, OrgID: access.OrgID()}).First(&component).Error; err != nil {
		return nil, err
	}
	appCfg, err := s.appsHelpers.GetComponentAppConfigInTx(ctx, access.DB(), component.AppID, "")
	if err != nil {
		return nil, err
	}

	if !generics.SliceContains(component.ID, appCfg.ComponentIDs) {
		return nil, stderr.ErrNotFound{
			Err:         errors.New("component does not belong to a current app config"),
			Description: "Component is not in the active app configuration",
		}
	}

	cdo, err := s.appsHelpers.GetComponentDependentsInTx(ctx, access.DB(), appCfg.ID, component.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component dependents")
	}

	comps := []app.Component{}
	if err := query.WithContext(ctx).Where(&app.Component{AppID: component.AppID}).Where("components.id IN ?", cdo).Order("components.id").Find(&comps).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get components")
	}
	byID := make(map[string]app.Component, len(comps))
	for _, comp := range comps {
		byID[comp.ID] = comp
	}
	ordered := make([]app.Component, 0, len(comps))
	for _, id := range cdo {
		if comp, ok := byID[id]; ok {
			ordered = append(ordered, comp)
		}
	}
	return ordered, nil
}
