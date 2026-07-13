package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/schema"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID						GetConfigSchema
// @Summary				Get jsonschema for config file (deprecated query form)
// @Description.markdown	config_schema.md
// @Tags					general
// @Accept					json
// @Param			type query	string	false	"return a schema for a source file"
// @Param			source query	string	false	"deprecated alias for type; responses include a Deprecation header when used"
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/general/config-schema [GET]
//
// Deprecated: all ?type=X share one URL path, so path-keyed schema caches
// collapse them. Use GetConfigSchemaByType.
func (s *service) GetConfigSchema(ctx *gin.Context) {
	typ := ctx.DefaultQuery("type", "")
	deprecatedSource := false
	if typ == "" {
		typ = ctx.DefaultQuery("source", "")
		deprecatedSource = typ != ""
	}
	if typ == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("type query parameter is required"),
			Description: "type query parameter is required",
		})
		return
	}

	if deprecatedSource {
		ctx.Header("Deprecation", "true")
		ctx.Header("Warning", `299 - "the source query parameter is deprecated; use type instead"`)
	}

	s.respondWithConfigSchema(ctx, typ, deprecatedSource)
}

// @ID						GetConfigSchemaByType
// @Summary				Get jsonschema for a config file type
// @Description.markdown	config_schema.md
// @Tags					general
// @Accept					json
// @Param			type path	string	true	"config file type, e.g. sandbox, terraform, action"
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/general/config-schema/{type} [GET]
func (s *service) GetConfigSchemaByType(ctx *gin.Context) {
	typ := ctx.Param("type")
	if typ == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("type path parameter is required"),
			Description: "type path parameter is required",
		})
		return
	}

	s.respondWithConfigSchema(ctx, typ, false)
}

// respondWithConfigSchema writes the schema for typ, stamping its $id with the
// exact fetch URL so tools that key by $id resolve each type distinctly.
func (s *service) respondWithConfigSchema(ctx *gin.Context, typ string, deprecatedSource bool) {
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

	schm.ID = jsonschema.ID(schemaFetchURL(ctx))

	if deprecatedSource {
		schm.Comments = "the source query parameter is deprecated; update the #:schema URL to use type instead"
	}

	ctx.JSON(http.StatusOK, schm)
}

// schemaFetchURL reconstructs the absolute URL used to fetch this schema, for
// use as its $id.
func schemaFetchURL(ctx *gin.Context) string {
	scheme := "https"
	if proto := ctx.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if ctx.Request.TLS == nil {
		scheme = "http"
	}

	url := scheme + "://" + ctx.Request.Host + ctx.Request.URL.Path
	if ctx.Request.URL.RawQuery != "" {
		url += "?" + ctx.Request.URL.RawQuery
	}
	return url
}
