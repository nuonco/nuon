package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/config/schema"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
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
	if typ == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("type query parameter is required"),
			Description: "type query parameter is required",
		})
		return
	}

	schm, err := schema.LookupSchemaType(typ)
	if err != nil {
		ctx.Error(err)
		return
	}

	if schm == nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unknown schema type: %s", typ),
			Description: "the provided type does not match any known schema types",
		})
		return
	}

	ctx.JSON(http.StatusOK, schm)
}
