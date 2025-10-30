package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/robfig/cron"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateHelmComponentConfigRequest struct {
	basicVCSConfigRequest

	Values        map[string]*string `json:"values,omitempty" validate:"required"`
	ValuesFiles   []string           `json:"values_files,omitempty"`
	ChartName     string             `json:"chart_name,omitempty" validate:"required,dns_rfc1035_label,min=5,max=62"`
	Namespace     string             `json:"namespace,omitempty"`
	StorageDriver string             `json:"storage_driver,omitempty"`
	TakeOwnership bool               `json:"take_ownership,omitempty"`

	AppConfigID string `json:"app_config_id"`

	Dependencies  []string `json:"dependencies"`
	References    []string `json:"references"`
	Checksum      string   `json:"checksum"`
	DriftSchedule *string  `json:"drift_schedule,omitempty"`
}

func (c *CreateHelmComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		return err
	}
	return nil
}

// @ID						CreateAppHelmComponentConfig
// @Summary				create a helm component config
// @Description.markdown	create_helm_component_config.md
// @Param					req				body	CreateHelmComponentConfigRequest	true	"Input"
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
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.HelmComponentConfig
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/helm [POST]
func (s *service) CreateAppHelmComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	// reuse the same logic as non-app scoped endpoint
	s.CreateHelmComponentConfig(ctx)
}

// @ID						CreateHelmComponentConfig
// @Summary				create a helm component config
// @Description.markdown	create_helm_component_config.md
// @Param					req				body	CreateHelmComponentConfigRequest	true	"Input"
// @Param					component_id	path	string								true	"component ID"
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
// @Success				201	{object}	app.HelmComponentConfig
// @Router					/v1/components/{component_id}/configs/helm [POST]
func (s *service) CreateHelmComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

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

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type: signals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:          signals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeHelmChart,
	})
	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createHelmComponentConfig(ctx context.Context, cmpID string, req *CreateHelmComponentConfigRequest) (*app.HelmComponentConfig, error) {
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
		return nil, fmt.Errorf("invalid connected github vcs config: %w", err)
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid public vcs config: %w", err)
	}

	cfg := app.HelmComponentConfig{
		PublicGitVCSConfig:       publicGitVCSConfig,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		HelmConfig: &app.HelmConfig{
			ChartName:     req.ChartName,
			Namespace:     req.Namespace,
			StorageDriver: req.StorageDriver,
			Values:        req.Values,
			ValuesFiles:   req.ValuesFiles,
			TakeOwnership: req.TakeOwnership,
		},
	}
	componentConfigConnection := app.ComponentConfigConnection{
		HelmComponentConfig:    &cfg,
		ComponentID:            parentCmp.ID,
		AppConfigID:            req.AppConfigID,
		ComponentDependencyIDs: pq.StringArray(depIDs),
		References:             pq.StringArray(req.References),
		Checksum:               req.Checksum,
	}
	if req.DriftSchedule != nil {
		_, err := cron.ParseStandard(*req.DriftSchedule)
		if err != nil {
			return nil, fmt.Errorf("invalid drift schedule: must be a valid cron expression: %s . Error: %s", *req.DriftSchedule, err.Error())
		}
		componentConfigConnection.DriftSchedule = *req.DriftSchedule
	}

	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create helm component config connection: %w", res.Error)
	}

	return &cfg, nil
}
