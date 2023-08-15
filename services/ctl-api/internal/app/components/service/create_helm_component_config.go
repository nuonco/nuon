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

type CreateHelmComponentConfigRequest struct {
	basicVCSConfigRequest

	Values    map[string]*string `json:"values,omitempty" validate:"required"`
	ChartName string             `json:"chart_name,omitempty" validate:"required"`
}

func (c *CreateHelmComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/components

// Create a helm component config
// @Summary create a helm component config
// @Schemes
// @Description create a helm component config.
// @Param req body CreateHelmComponentConfigRequest true "Input"
// @Param component_id path string component_id "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 201 {object} app.ComponentConfigConnection
// @Router /v1/components/{component_id}/configs/helm [POST]
func (s *service) CreateHelmComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

	var req CreateHelmComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createHelmComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) createHelmComponentConfig(ctx context.Context, cmpID string, req *CreateHelmComponentConfigRequest) (*app.HelmComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	// build component config
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(parentCmp)
	if err != nil {
		return nil, fmt.Errorf("invalid connected github vcs config: %w", err)
	}

	cfg := app.HelmComponentConfig{
		PublicGitVCSConfig:       req.publicGitVCSConfig(),
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		Values:                   pgtype.Hstore(req.Values),
		ChartName:                req.ChartName,
	}
	componentConfigConnection := app.ComponentConfigConnection{
		HelmComponentConfig: &cfg,
		ComponentID:         parentCmp.ID,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create helm component config connection: %w", res.Error)
	}

	return &cfg, nil
}
