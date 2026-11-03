package service

import (
	"github.com/gin-gonic/gin"
)

// @ID						GetBuild
// @Summary				get a build
// @Description.markdown	get_component_build.md
// @Param					build_id	path	string	true	"build ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.ComponentBuild
// @Router					/v1/components/builds/{build_id} [GET]
func (s *service) GetBuild(ctx *gin.Context) (any, error) {
	return s.getComponentBuild(ctx.Request.Context(), "", ctx.Param("build_id"))
}
