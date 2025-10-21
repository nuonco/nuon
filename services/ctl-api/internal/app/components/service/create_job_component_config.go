package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

type CreateJobComponentConfigRequest struct {
	ImageURL string             `json:"image_url" validate:"required"`
	Tag      string             `json:"tag" validate:"required"`
	Cmd      []string           `json:"cmd"`
	EnvVars  map[string]*string `json:"env_vars"`
	Args     []string           `json:"args"`

	AppConfigID string   `json:"app_config_id"`
	References  []string `json:"references"`
	Checksum    string   `json:"checksum"`
}

func (c *CreateJobComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateAppJobComponentConfig
// @Summary				create a job component config
// @Description.markdown	create_job_component_config.md
// @Param					req				body	CreateJobComponentConfigRequest	true	"Input"
// @Param					component_id	path	string							true	"component ID"
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
// @Success				201	{object}	app.JobComponentConfig
// @Router					/v1/components/{component_id}/configs/job [POST]
func (s *service) CreateAppJobComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	// reuse the same logic as non-app scoped endpoint
	s.CreateJobComponentConfig(ctx)
}

// @ID						CreateJobComponentConfig
// @Summary				create a job component config
// @Description.markdown	create_job_component_config.md
// @Param					req				body	CreateJobComponentConfigRequest	true	"Input"
// @Param					component_id	path	string							true	"component ID"
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
// @Success				201	{object}	app.JobComponentConfig
// @Router					/v1/components/{component_id}/configs/job [POST]
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

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type: signals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:          signals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeJob,
	})
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
		AppConfigID:        req.AppConfigID,
		References:         pq.StringArray(req.References),
		Checksum:           req.Checksum,
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create job component config connection: %w", res.Error)
	}

	return &cfg, nil
}
