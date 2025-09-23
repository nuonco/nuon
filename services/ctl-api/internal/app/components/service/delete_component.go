package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

// @ID						DeleteComponent
// @Summary				delete a component
// @Description.markdown	delete_component.md
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
// @Router					/v1/components/{component_id} [DELETE]
func (s *service) DeleteComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	err := s.appsHelpers.DeleteAppComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, componentID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}
