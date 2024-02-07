package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID GetComponentConfigs
// @Summary	get all configs for a component
// @Description.markdown	get_component_configs.md
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
// @Success		200				{array}		app.ComponentConfigConnection
// @Router			/v1/components/{component_id}/configs [GET]
func (s *service) GetComponentConfigs(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	comp, err := s.helpers.GetComponent(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, comp.ComponentConfigs)
}
