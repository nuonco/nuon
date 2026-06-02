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
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppSandboxConfigRequest struct {
	vcshelpers.VCSConfigRequest

	Type                         string  `json:"type,omitempty"`
	TerraformVersion             string  `json:"terraform_version,omitempty"`
	Runtime                      string  `json:"runtime,omitempty"`
	PulumiVersion                string  `json:"pulumi_version,omitempty"`
	DriftSchedule                *string `json:"drift_schedule,omitempty" validate:"omitempty,cron_schedule"`
	MaxAutoRetries               *int    `json:"max_auto_retries,omitempty"`
	SkipNoops                    *bool   `json:"skip_noops,omitempty"`
	AutoApproveOnPoliciesPassing *bool   `json:"auto_approve_on_policies_passing,omitempty"`

	VariablesFiles []string           `json:"variables_files,omitempty"`
	Variables      map[string]*string `json:"variables" validate:"required"`
	EnvVars        map[string]*string `json:"env_vars" validate:"required"`
	PulumiConfig   map[string]*string `json:"pulumi_config,omitempty"`

	OperationRoles map[app.OperationType]*string `json:"operation_roles,omitempty"`

	References []string `json:"references"`

	AppConfigID string `json:"app_config_id"`
}

func (c *CreateAppSandboxConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	sandboxType := c.Type
	if sandboxType == "" {
		sandboxType = config.AppSandboxTypeTerraform
	}

	switch sandboxType {
	case config.AppSandboxTypeTerraform:
		if c.TerraformVersion == "" {
			return fmt.Errorf("terraform_version is required when type=terraform")
		}
	case config.AppSandboxTypePulumi:
		if c.Runtime == "" {
			return fmt.Errorf("runtime is required when type=pulumi")
		}
		if !slices.Contains(config.ValidPulumiRuntimes, c.Runtime) {
			return fmt.Errorf("invalid pulumi runtime: %s. Valid runtimes: %v", c.Runtime, config.ValidPulumiRuntimes)
		}
	default:
		return fmt.Errorf("invalid sandbox type: %s. Valid types: terraform, pulumi", sandboxType)
	}

	if c.OperationRoles != nil {
		for operation := range c.OperationRoles {
			if !slices.Contains(app.ValidOperations, operation) {
				return fmt.Errorf("invalid operation type: %s. Valid operations: %v", operation, app.ValidOperations)
			}
		}
	}

	if c.MaxAutoRetries != nil {
		if err := validateMaxAutoRetries(*c.MaxAutoRetries); err != nil {
			return err
		}
	}

	return nil
}

// @ID						CreateAppSandboxConfigV2
// @Summary				create an app sandbox config
// @Description.markdown	create_app_sandbox_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppSandboxConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppSandboxConfig
// @Router					/v1/apps/{app_id}/sandbox-configs [post]
func (s *service) CreateAppSandboxConfigV2(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSandboxConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sandboxConfig, err := s.createAppSandboxConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app sandbox config: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, sandboxConfig)
}

// @ID						CreateAppSandboxConfig
// @Summary				create an app sandbox config
// @Description.markdown	create_app_sandbox_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppSandboxConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppSandboxConfig
// @Router					/v1/apps/{app_id}/sandbox-config [post]
func (s *service) CreateAppSandboxConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSandboxConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sandboxConfig, err := s.createAppSandboxConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app sandbox config: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, sandboxConfig)
}

func (s *service) createAppSandboxConfig(ctx context.Context, appID string, req *CreateAppSandboxConfigRequest) (*app.AppSandboxConfig, error) {
	var parentApp app.App
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app sandbox: %w", res.Error)
	}

	sandboxType := req.Type
	if sandboxType == "" {
		sandboxType = config.AppSandboxTypeTerraform
	}
	if sandboxType == config.AppSandboxTypePulumi {
		enabled, err := s.featuresClient.OrgHasFeature(ctx, parentApp.OrgID, app.OrgFeaturePulumiSandbox)
		if err != nil {
			return nil, fmt.Errorf("unable to check pulumi-sandbox feature flag: %w", err)
		}
		if !enabled {
			return nil, stderr.NewInvalidRequest(fmt.Errorf("pulumi sandboxes are not enabled for this organization"))
		}
	}

	// Build VCS configs
	githubVCSConfig, err := s.vcsHelpers.BuildConnectedGithubVCSConfig(ctx, req.ConnectedGithubVCSConfig, parentApp.Org)
	if err != nil {
		return nil, fmt.Errorf("unable to create connected github vcs config: %w", err)
	}

	publicGitConfig, err := s.vcsHelpers.BuildPublicGitVCSConfig(ctx, req.PublicGitVCSConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get public git config: %w", err)
	}

	operationRoles := make(pgtype.Hstore)
	for operation, role := range req.OperationRoles {
		operationRoles[string(operation)] = role
	}

	appSandboxConfig := app.AppSandboxConfig{
		AppID:                        appID,
		AppConfigID:                  req.AppConfigID,
		PublicGitVCSConfig:           publicGitConfig,
		ConnectedGithubVCSConfig:     githubVCSConfig,
		Variables:                    pgtype.Hstore(req.Variables),
		EnvVars:                      pgtype.Hstore(req.EnvVars),
		VariablesFiles:               pq.StringArray(req.VariablesFiles),
		Type:                         sandboxType,
		TerraformVersion:             req.TerraformVersion,
		Runtime:                      req.Runtime,
		PulumiVersion:                req.PulumiVersion,
		PulumiConfig:                 pgtype.Hstore(req.PulumiConfig),
		References:                   pq.StringArray(req.References),
		OperationRoles:               operationRoles,
		MaxAutoRetries:               req.MaxAutoRetries,
		SkipNoops:                    req.SkipNoops,
		AutoApproveOnPoliciesPassing: req.AutoApproveOnPoliciesPassing,
	}

	if req.DriftSchedule != nil {
		appSandboxConfig.DriftSchedule = *req.DriftSchedule
	}

	res = s.db.WithContext(ctx).
		Create(&appSandboxConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app sandbox config: %w", res.Error)
	}

	return &appSandboxConfig, nil
}
