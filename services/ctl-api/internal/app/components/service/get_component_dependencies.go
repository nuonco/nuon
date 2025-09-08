package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetComponentDependencies
// @Summary				get a component's dependencies
// @Description.markdown	get_component_dependencies.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Component
// @Router					/v1/components/{component_id}/dependencies [get]
func (s *service) GetComponentDependencies(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
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

	appCfg, err := s.appsHelpers.GetAppLatestConfig(ctx, component.AppID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get app config"))
		return
	}

	depCmps, err := s.appsHelpers.GetComponentDependents(ctx, appCfg.ID, component.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component dependencies: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, depCmps)
}
