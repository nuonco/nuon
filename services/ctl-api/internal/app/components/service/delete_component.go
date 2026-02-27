package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						DeleteAppComponent
// @Summary				delete a component
// @Description.markdown	delete_component.md
// @Param					app_id			path	string	true	"app ID"
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
// @Success				200	{boolean}	true
// @Router					/v1/apps/{app_id}/components/{component_id} [DELETE]
func (s *service) DeleteAppComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")

	// Validate component belongs to org before deleting
	_, err = s.findComponent(ctx, org.ID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find component %s: %w", componentID, err))
		return
	}

	err = s.appsHelpers.DeleteAppComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, componentID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}

// @ID						DeleteComponent
// @Summary				delete a component
// @Description.markdown	delete_component.md
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
// @Success				200	{boolean}	true
// @Router					/v1/components/{component_id} [DELETE]
func (s *service) DeleteComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")

	// Validate component belongs to org before deleting
	_, err = s.findComponent(ctx, org.ID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find component %s: %w", componentID, err))
		return
	}

	err = s.appsHelpers.DeleteAppComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, componentID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}
