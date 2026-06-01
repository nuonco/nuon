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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreatePulumiComponentConfigRequest struct {
	basicVCSConfigRequest

	Runtime                      string             `json:"runtime" validate:"required"`
	Version                      string             `json:"version"`
	Config                       map[string]*string `json:"config" validate:"required"`
	EnvVars                      map[string]*string `json:"env_vars" validate:"required"`
	BuildTimeout                 string             `json:"build_timeout,omitempty"`
	DeployTimeout                string             `json:"deploy_timeout,omitempty"`
	MaxAutoRetries               *int               `json:"max_auto_retries,omitempty"`
	SkipNoops                    *bool              `json:"skip_noops,omitempty"`
	AutoApproveOnPoliciesPassing *bool              `json:"auto_approve_on_policies_passing,omitempty"`

	AppConfigID string `json:"app_config_id"`

	Dependencies   []string                      `json:"dependencies"`
	References     []string                      `json:"references"`
	Checksum       string                        `json:"checksum"`
	DriftSchedule  *string                       `json:"drift_schedule,omitempty" validate:"omitempty,cron_schedule"`
	OperationRoles map[app.OperationType]*string `json:"operation_roles,omitempty"`
}

var validPulumiRuntimes = []string{"go", "nodejs", "python", "dotnet", "java", "yaml"}

func (c *CreatePulumiComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if !slices.Contains(validPulumiRuntimes, c.Runtime) {
		return fmt.Errorf("invalid runtime %s: must be one of %v", c.Runtime, validPulumiRuntimes)
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

// @ID						CreateAppPulumiComponentConfig
// @Summary				create a pulumi component config
// @Param					req				body	CreatePulumiComponentConfigRequest	true	"Input"
// @Param					app_id			path	string								true	"app ID"
// @Param					component_id	path	string								true	"component ID"
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
// @Success				201	{object}	app.PulumiComponentConfig
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/pulumi [POST]
func (s *service) CreateAppPulumiComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	s.CreatePulumiComponentConfig(ctx)
}

// @ID						CreatePulumiComponentConfig
// @Summary				create a pulumi component config
// @Param					req				body	CreatePulumiComponentConfigRequest	true	"Input"
// @Param					component_id	path	string								true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated 	  true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.PulumiComponentConfig
// @Router					/v1/components/{component_id}/configs/pulumi [POST]
func (s *service) CreatePulumiComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreatePulumiComponentConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	cfg, err := s.createPulumiComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	if err := s.onConfigCreated(ctx, cmpID, app.ComponentTypePulumi); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createPulumiComponentConfig(ctx context.Context, cmpID string, req *CreatePulumiComponentConfigRequest) (*app.PulumiComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	depIDs, err := s.helpers.GetComponentIDs(ctx, parentCmp.AppID, req.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("unable to get component ids: %w", err)
	}

	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid connected github config: %w", err)
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid public vcs config: %w", err)
	}

	cfg := app.PulumiComponentConfig{
		Runtime:                  req.Runtime,
		Version:                  req.Version,
		PublicGitVCSConfig:       publicGitVCSConfig,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		Config:                   pgtype.Hstore(req.Config),
		EnvVars:                  pgtype.Hstore(req.EnvVars),
	}

	var operationRoles pgtype.Hstore
	if req.OperationRoles != nil {
		operationRoles = make(pgtype.Hstore)
		for operation, role := range req.OperationRoles {
			operationRoles[string(operation)] = role
		}
	}

	componentConfigConnection := app.ComponentConfigConnection{
		PulumiComponentConfig:        &cfg,
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
	if req.DriftSchedule != nil {
		componentConfigConnection.DriftSchedule = *req.DriftSchedule
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create pulumi component config connection: %w", res.Error)
	}

	err = s.helpers.UpdateComponentType(ctx, cmpID, app.ComponentTypePulumi)
	if err != nil {
		return nil, fmt.Errorf("unable to update component type: %w", err)
	}

	return &cfg, nil
}
