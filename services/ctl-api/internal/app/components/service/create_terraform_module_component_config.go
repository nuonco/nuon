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
}

func (c *CreateTerraformModuleComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/components

// Create a terraform component config
// @Summary create a terraform component config
// @Schemes
// @Description create a terraform component config.
// @Param req body CreateTerraformModuleComponentConfigRequest true "Input"
// @Param component_id path string component_id "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 201 {object} app.ComponentConfigConnection
// @Router /v1/components/{component_id}/configs/terraform-module [POST]
func (s *service) CreateTerraformModuleComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

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

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) createTerraformModuleComponentConfig(ctx context.Context, cmpID string, req *CreateTerraformModuleComponentConfigRequest) (*app.TerraformModuleComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	// build component config
	cfg := app.TerraformModuleComponentConfig{
		Version:                  req.Version,
		PublicGitVCSConfig:       req.publicGitVCSConfig(),
		ConnectedGithubVCSConfig: req.connectedGithubVCSConfig(parentCmp),
		Variables:                pgtype.Hstore(req.Variables),
	}

	componentConfigConnection := app.ComponentConfigConnection{
		TerraformModuleComponentConfig: &cfg,
		ComponentID:                    parentCmp.ID,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create terraform component config connection: %w", res.Error)
	}

	return &cfg, nil
}
