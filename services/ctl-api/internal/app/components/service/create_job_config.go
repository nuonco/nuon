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

type CreateJobComponentConfigRequest struct {
	ImageURL string             `json:"image_url" validate:"required"`
	Tag      string             `json:"tag" validate:"required"`
	Cmd      []string           `json:"cmd"`
	EnvVars  map[string]*string `json:"env_vars"`
	Args     []string           `json:"args"`
}

func (c *CreateJobComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

//	@BasePath	/v1/components

// Create a job component config
//
//	@Summary	create a job component config
//	@Schemes
//	@Description	create a job component config.
//	@Param			req				body	CreateJobComponentConfigRequest	true	"Input"
//	@Param			component_id	path	string							true	"component ID"
//	@Tags			components
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		201				{object}	app.JobComponentConfig
//	@Router			/v1/components/{component_id}/configs/job [POST]
func (s *service) CreateJobComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateJobComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createJobComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createJobComponentConfig(ctx context.Context, cmpID string, req *CreateJobComponentConfigRequest) (*app.JobComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	// build component config
	cfg := app.JobComponentConfig{
		ImageURL: req.ImageURL,
		Tag:      req.Tag,
		Cmd:      req.Cmd,
		EnvVars:  pgtype.Hstore(req.EnvVars),
		Args:     req.Args,
	}

	componentConfigConnection := app.ComponentConfigConnection{
		JobComponentConfig: &cfg,
		ComponentID:        parentCmp.ID,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create job component config connection: %w", res.Error)
	}

	return &cfg, nil
}
