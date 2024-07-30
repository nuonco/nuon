package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/invopop/jsonschema"

	"github.com/powertoolsdev/mono/pkg/config/schema"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID GetConfigSchema
// @Summary	Get jsonschema for config file
// @Description.markdown	config_schema.md
// @Tags			general
// @Accept			json
// @Param   source query string false	"return a schema for a source file"
// @Param   flat query string false	"return a flat schema for the full app"
// @Produce		json
// @Failure		400	{object}	stderr.ErrResponse
// @Failure		401	{object}	stderr.ErrResponse
// @Failure		403	{object}	stderr.ErrResponse
// @Failure		404	{object}	stderr.ErrResponse
// @Failure		500	{object}	stderr.ErrResponse
// @Success		200	{object} interface{}
// @Router			/v1/general/config-schema [GET]
func (s *service) GetConfigSchema(ctx *gin.Context) {
	src := ctx.DefaultQuery("source", "")
	flat := ctx.DefaultQuery("flat", "")

	var fn func() (*jsonschema.Schema, error)

	mapping := make(map[[2]string]func() (*jsonschema.Schema, error))
	mapping[[2]string{"", "inputs"}] = schema.InputsSourceSchema
	mapping[[2]string{"", "installer"}] = schema.InstallerSourceSchema
	mapping[[2]string{"", "runner"}] = schema.RunnerSourceSchema
	mapping[[2]string{"", "sandbox"}] = schema.SandboxSourceSchema
	mapping[[2]string{"", "docker_build"}] = schema.DockerBuildComponent
	mapping[[2]string{"", "helm"}] = schema.HelmComponent
	mapping[[2]string{"", "terraform"}] = schema.TerraformComponent
	mapping[[2]string{"", "job"}] = schema.JobComponent
	mapping[[2]string{"", "container_image"}] = schema.ContainerImageComponent
	mapping[[2]string{"true", ""}] = schema.AppSchema
	mapping[[2]string{"", ""}] = schema.AppSchemaSources

	fn, ok := mapping[[2]string{flat, src}]
	if !ok {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid source,flat option %s %s", src, flat),
			Description: "Invalid source option",
		})
		return
	}

	schm, err := fn()
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schm)
}
