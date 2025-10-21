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
)

type CreateKubernetesManifestComponentConfigRequest struct {
	AppConfigID string `json:"app_config_id"`

	References   []string `json:"references"`
	Checksum     string   `json:"checksum"`
	Dependencies []string `json:"dependencies"`

	Manifest      string  `json:"manifest"`
	Namespace     string  `json:"namespace"`
	DriftSchedule *string `json:"drift_schedule,omitempty"`
}

func (c *CreateKubernetesManifestComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateAppKubernetesManifestComponentConfig
// @Summary					create a kubernetes manifest component config
// @Description.markdown	create_kubernetes_manifest_component_config.md
// @Param					req				body	CreateKubernetesManifestComponentConfigRequest	true	"Input"
// @Param					component_id	path	string							true	"component ID"
// @Tags					components
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.KubernetesManifestComponentConfig
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/kubernetes-manifest [POST]
func (s *service) CreateAppKubernetesManifestComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	// reuse the same logic as non-app scoped endpoint
	s.CreateKubernetesManifestComponentConfig(ctx)
}

// @ID						CreateKubernetesManifestComponentConfig
// @Summary					create a kubernetes manifest component config
// @Description.markdown	create_kubernetes_manifest_component_config.md
// @Param					req				body	CreateKubernetesManifestComponentConfigRequest	true	"Input"
// @Param					component_id	path	string							true	"component ID"
// @Tags					components
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.KubernetesManifestComponentConfig
// @Router					/v1/components/{component_id}/configs/kubernetes-manifest [POST]
func (s *service) CreateKubernetesManifestComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateKubernetesManifestComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createKubernetesManifestComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	// sk: this triggers queue build
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type: signals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:          signals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeKubernetesManifest,
	})
	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createKubernetesManifestComponentConfig(
	ctx context.Context, cmpID string, req *CreateKubernetesManifestComponentConfigRequest,
) (*app.KubernetesManifestComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	depIDs, err := s.helpers.GetComponentIDs(ctx, parentCmp.AppID, req.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}

	// build component config
	cfg := app.KubernetesManifestComponentConfig{
		Manifest:  req.Manifest,
		Namespace: req.Namespace,
	}

	componentConfigConnection := app.ComponentConfigConnection{
		KubernetesManifestComponentConfig: &cfg,
		ComponentID:                       parentCmp.ID,
		AppConfigID:                       req.AppConfigID,
		References:                        pq.StringArray(req.References),
		Checksum:                          req.Checksum,
		ComponentDependencyIDs:            pq.StringArray(depIDs),
	}

	if req.DriftSchedule != nil {
		_, err := cron.ParseStandard(*req.DriftSchedule)
		if err != nil {
			return nil, fmt.Errorf("invalid drift schedule: must be a valid cron expression: %s . Error: %s", *req.DriftSchedule, err.Error())
		}
		componentConfigConnection.DriftSchedule = *req.DriftSchedule

	}

	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create kubernetes component config connection: %w", res.Error)
	}

	return &cfg, nil
}
