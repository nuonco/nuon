package service

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateDockerBuildComponentConfigRequest struct {
	basicVCSConfigRequest

	Dockerfile                   string             `json:"dockerfile" validate:"required"`
	Target                       string             `json:"target"`
	BuildArgs                    []string           `json:"build_args"`
	EnvVars                      map[string]*string `json:"env_vars"`
	BuildTimeout                 string             `json:"build_timeout,omitempty"`  // Duration string for build operations (e.g., "30m", "1h")
	DeployTimeout                string             `json:"deploy_timeout,omitempty"` // Duration string for deploy operations (e.g., "30m", "1h")
	MaxAutoRetries               *int               `json:"max_auto_retries,omitempty"`
	SkipNoops                    *bool              `json:"skip_noops,omitempty"`
	Toggleable                   *bool              `json:"toggleable,omitempty"`
	DefaultEnabled               *bool              `json:"default_enabled,omitempty"`
	AutoApproveOnPoliciesPassing *bool              `json:"auto_approve_on_policies_passing,omitempty"`
	AppConfigID                  string             `json:"app_config_id"`

	Dependencies   []string                      `json:"dependencies"`
	References     []string                      `json:"references"`
	Checksum       string                        `json:"checksum"`
	OperationRoles map[app.OperationType]*string `json:"operation_roles,omitempty"`
}

func (c *CreateDockerBuildComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.OperationRoles != nil {
		for operation := range c.OperationRoles {
			if !slices.Contains(app.ValidOperations, operation) {
				return fmt.Errorf("invalid operation type: %s. Valid operations: %v", operation, app.ValidOperations)
			}
		}
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		return err
	}

	if c.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(c.BuildTimeout); err != nil {
			return err
		}
	}
	if c.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(c.DeployTimeout); err != nil {
			return err
		}
	}
	if c.MaxAutoRetries != nil {
		if err := validation.ValidateMaxAutoRetries(*c.MaxAutoRetries); err != nil {
			return err
		}
	}

	return nil
}

// @ID						CreateAppDockerBuildComponentConfig
// @Summary				create a docker build component config
// @Description.markdown	create_docker_build_component_config.md
// @Param					req				body	CreateDockerBuildComponentConfigRequest	true	"Input"
// @Param					app_id			path	string									true	"app ID"
// @Param					component_id	path	string									true	"component ID"
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
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.DockerBuildComponentConfig
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/docker-build [POST]
func (s *service) CreateAppDockerBuildComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	// reuse the same logic as non-app scoped endpoint
	s.CreateDockerBuildComponentConfig(ctx)
}

// @ID						CreateDockerBuildComponentConfig
// @Summary				create a docker build component config
// @Description.markdown	create_docker_build_component_config.md
// @Param					req				body	CreateDockerBuildComponentConfigRequest	true	"Input"
// @Param					component_id	path	string									true	"component ID"
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
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.DockerBuildComponentConfig
// @Router					/v1/components/{component_id}/configs/docker-build [POST]
func (s *service) CreateDockerBuildComponentConfig(ctx *gin.Context) {
	ctx.Error(componenthelpers.DockerBuildUnsupported())
}
