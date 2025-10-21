package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

type CreateDockerBuildComponentConfigRequest struct {
	basicVCSConfigRequest

	Dockerfile  string             `json:"dockerfile" validate:"required"`
	Target      string             `json:"target"`
	BuildArgs   []string           `json:"build_args"`
	EnvVars     map[string]*string `json:"env_vars"`
	AppConfigID string             `json:"app_config_id"`

	Dependencies []string `json:"dependencies"`
	References   []string `json:"references"`
	Checksum     string   `json:"checksum"`
}

func (c *CreateDockerBuildComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		return err
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
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.DockerBuildComponentConfig
// @Router					/v1/components/{component_id}/configs/docker-build [POST]
func (s *service) CreateDockerBuildComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateDockerBuildComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createDockerBuildComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type: signals.OperationConfigCreated,
	})

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:          signals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeDockerBuild,
	})

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

	componentConfigConnection := app.ComponentConfigConnection{
		DockerBuildComponentConfig: &cfg,
		ComponentID:                parentCmp.ID,
		AppConfigID:                req.AppConfigID,
		ComponentDependencyIDs:     pq.StringArray(depIDs),
		References:                 pq.StringArray(req.References),
		Checksum:                   req.Checksum,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create docker build component config connection: %w", res.Error)
	}

	return &cfg, nil
}
