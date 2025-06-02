package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/config/schema"
)

// @ID						GetConfigSchema
// @Summary				Get jsonschema for config file
// @Description.markdown	config_schema.md
// @Tags					general
// @Accept					json
// @Param			type query	string	false	"return a schema for a source file"
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/general/config-schema [GET]
func (s *service) GetConfigSchema(ctx *gin.Context) {
	typ := ctx.DefaultQuery("type", "")

	schm, err := schema.LookupSchemaType(typ)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schm)
}
