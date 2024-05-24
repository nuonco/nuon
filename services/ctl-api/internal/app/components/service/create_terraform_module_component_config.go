package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateTerraformModuleComponentConfigRequest struct {
	basicVCSConfigRequest

	Version   string             `json:"version"`
	Variables map[string]*string `json:"variables" validate:"required"`
	EnvVars   map[string]*string `json:"env_vars" validate:"required"`
}

func (c *CreateTerraformModuleComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateTerraformModuleComponentConfig
// @Summary	create a terraform component config
// @Description.markdown	create_terraform_component_config.md
// @Param			req				body	CreateTerraformModuleComponentConfigRequest	true	"Input"
// @Param			component_id	path	string										true	"component ID"
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.TerraformModuleComponentConfig
// @Router			/v1/components/{component_id}/configs/terraform-module [POST]
func (s *service) CreateTerraformModuleComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateTerraformModuleComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createTerraformModuleComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	s.hooks.ConfigCreated(ctx, cmpID)
	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createTerraformModuleComponentConfig(ctx context.Context, cmpID string, req *CreateTerraformModuleComponentConfigRequest) (*app.TerraformModuleComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	// build component config
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid connected github config: %w", err)
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid public vcs config: %w", err)
	}

	cfg := app.TerraformModuleComponentConfig{
		Version:                  req.Version,
		PublicGitVCSConfig:       publicGitVCSConfig,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		Variables:                pgtype.Hstore(req.Variables),
		EnvVars:                  pgtype.Hstore(req.EnvVars),
	}

	componentConfigConnection := app.ComponentConfigConnection{
		Version:                        parentCmp.ConfigVersions + 1,
		TerraformModuleComponentConfig: &cfg,
		ComponentID:                    parentCmp.ID,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create terraform component config connection: %w", res.Error)
	}

	return &cfg, nil
}
