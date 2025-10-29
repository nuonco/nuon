package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
)

type CreateAppSandboxConfigRequest struct {
	basicVCSConfigRequest

	TerraformVersion string  `json:"terraform_version" validate:"required"`
	DriftSchedule    *string `json:"drift_schedule,omitempty"`

	VariablesFiles []string           `json:"variables_files,omitempty"`
	Variables      map[string]*string `json:"variables" validate:"required"`
	EnvVars        map[string]*string `json:"env_vars" validate:"required"`

	References []string `json:"references"`

	AppConfigID string `json:"app_config_id"`
}

func (c *CreateAppSandboxConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
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
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppSandboxConfig
// @Router					/v1/apps/{app_id}/sandbox-configs [post]
func (s *service) CreateAppSandboxConfigV2(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSandboxConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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

	s.evClient.Send(ctx, appID, &signals.Signal{
		Type:               signals.OperationUpdateSandbox,
		AppSandboxConfigID: sandboxConfig.ID,
	})
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
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppSandboxConfig
// @Router					/v1/apps/{app_id}/sandbox-config [post]
func (s *service) CreateAppSandboxConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSandboxConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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

	s.evClient.Send(ctx, appID, &signals.Signal{
		Type:               signals.OperationUpdateSandbox,
		AppSandboxConfigID: sandboxConfig.ID,
	})
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

	// build the app sandbox config
	githubVCSConfig, err := req.connectedGithubVCSConfig(ctx, &parentApp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("unable to create connected github vcs config: %w", err)
	}

	publicGitConfig, err := req.publicGitVCSConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get public git config: %w", err)
	}

	appSandboxConfig := app.AppSandboxConfig{
		AppID:                    appID,
		AppConfigID:              req.AppConfigID,
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
		Variables:                pgtype.Hstore(req.Variables),
		EnvVars:                  pgtype.Hstore(req.EnvVars),
		VariablesFiles:           pq.StringArray(req.VariablesFiles),
		TerraformVersion:         req.TerraformVersion,
		References:               pq.StringArray(req.References),
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
