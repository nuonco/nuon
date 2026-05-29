package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
		if err := validateBuildTimeout(c.BuildTimeout); err != nil {
			return err
		}
	}
	if c.DeployTimeout != "" {
		if err := validateDeployTimeout(c.DeployTimeout); err != nil {
			return err
		}
	}
	if c.MaxAutoRetries != nil {
		if err := validateMaxAutoRetries(*c.MaxAutoRetries); err != nil {
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
	cmpID := ctx.Param("component_id")

	// docker_build runs kaniko in-process inside the build runner pod,
	// which mutates the runner container's rootfs (removes /usr/bin/git
	// and other tooling). The terraform-provider-mirror feature relies
	// on `terraform get` finding git on PATH at build time, so the two
	// features are mutually incompatible inside the same pod. Refuse to
	// create new docker_build configs for orgs that have the mirror
	// flag enabled. Existing configs are left in the DB; their builds
	// will fail at planning time via createDockerBuildPlan.
	mirrorEnabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureTerraformProviderMirror)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check terraform provider mirror feature flag: %w", err))
		return
	}
	if mirrorEnabled {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("docker_build components are not supported when the %q feature is enabled", app.OrgFeatureTerraformProviderMirror),
			Description: fmt.Sprintf(
				"This organization has the %q feature enabled, which is incompatible with docker_build components. "+
					"Use a container_image component to reference a pre-built image, or contact support to disable the feature.",
				app.OrgFeatureTerraformProviderMirror,
			),
			Code: "docker_build_incompatible_with_provider_mirror",
		})
		return
	}

	var req CreateDockerBuildComponentConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	cfg, err := s.createDockerBuildComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	if err := s.onConfigCreated(ctx, cmpID, app.ComponentTypeDockerBuild); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createDockerBuildComponentConfig(ctx context.Context, cmpID string, req *CreateDockerBuildComponentConfigRequest) (*app.DockerBuildComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	depIDs, err := s.helpers.GetComponentIDs(ctx, parentCmp.AppID, req.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}

	// build component config
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid github vcs config: %w", err)
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid public vcs config: %w", err)
	}

	cfg := app.DockerBuildComponentConfig{
		PublicGitVCSConfig:       publicGitVCSConfig,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,

		Dockerfile: req.Dockerfile,
		Target:     req.Target,
		BuildArgs:  req.BuildArgs,
		EnvVars:    pgtype.Hstore(req.EnvVars),
	}

	operationRoles := make(pgtype.Hstore)
	for operation, role := range req.OperationRoles {
		operationRoles[string(operation)] = role
	}

	componentConfigConnection := app.ComponentConfigConnection{
		DockerBuildComponentConfig:   &cfg,
		ComponentID:                  parentCmp.ID,
		AppConfigID:                  req.AppConfigID,
		ComponentDependencyIDs:       pq.StringArray(depIDs),
		References:                   pq.StringArray(req.References),
		Checksum:                     req.Checksum,
		BuildTimeout:                 req.BuildTimeout,
		DeployTimeout:                req.DeployTimeout,
		MaxAutoRetries:               req.MaxAutoRetries,
		SkipNoops:                    req.SkipNoops,
		AutoApproveOnPoliciesPassing: req.AutoApproveOnPoliciesPassing,
		OperationRoles:               operationRoles,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create docker build component config connection: %w", res.Error)
	}

	return &cfg, nil
}
