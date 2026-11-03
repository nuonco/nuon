package service

import (
	"github.com/gin-gonic/gin"
)

// @ID						GetAppComponentDependencies
// @Summary				get a component's dependencies
// @Description.markdown	get_component_dependencies.md
// @Param					app_id			path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Component
// @Router					/v1/apps/{app_id}/components/{component_id}/dependencies [get]
// XXX: This endpoint calls GetComponentDependents which does a BFS following outgoing edges
// (dependency → dependent). This means it returns the component's *dependents* (children),
// not its *dependencies* (parents). To return actual dependencies, the BFS would need to
// traverse incoming edges instead.
func (s *service) GetAppComponentDependencies(ctx *gin.Context) (any, error) {
	return s.GetComponentDependencies(ctx)
}

// @ID						GetComponentDependencies
// @Summary				get a component's dependencies
// @Description.markdown	get_component_dependencies.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Component
// @Router					/v1/components/{component_id}/dependencies [get]
func (s *service) GetComponentDependencies(ctx *gin.Context) (any, error) {
	components, err := s.readComponentDependents(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(components))
	for _, component := range components {
		ids = append(ids, component.ID)
	}
	return ids, nil
}
