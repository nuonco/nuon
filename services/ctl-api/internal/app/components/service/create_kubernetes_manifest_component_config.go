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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateKubernetesManifestComponentConfigRequest struct {
	AppConfigID string `json:"app_config_id"`

	References   []string `json:"references"`
	Checksum     string   `json:"checksum"`
	Dependencies []string `json:"dependencies"`

	// Inline manifest (mutually exclusive with Kustomize)
	Manifest      string  `json:"manifest,omitempty"`
	Namespace     string  `json:"namespace"`
	DriftSchedule *string `json:"drift_schedule,omitempty"`

	// Kustomize configuration (mutually exclusive with Manifest)
	Kustomize *KustomizeConfigRequest `json:"kustomize,omitempty"`

	// VCS configuration for kustomize sources
	basicVCSConfigRequest
}

// KustomizeConfigRequest defines kustomize options in API requests
type KustomizeConfigRequest struct {
	Path           string   `json:"path"`
	Patches        []string `json:"patches,omitempty"`
	EnableHelm     bool     `json:"enable_helm,omitempty"`
	LoadRestrictor string   `json:"load_restrictor,omitempty"`
}

func (c *CreateKubernetesManifestComponentConfigRequest) Validate(v *validator.Validate) error {
	// Normalize: treat kustomize with empty path as nil
	// This handles the case where go-swagger client sends {"kustomize": {"path": null}}
	if c.Kustomize != nil && c.Kustomize.Path == "" {
		c.Kustomize = nil
	}

	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	// Exactly one of manifest or kustomize must be set
	hasManifest := c.Manifest != ""
	hasKustomize := c.Kustomize != nil

	if !hasManifest && !hasKustomize {
		return errors.New("one of 'manifest' or 'kustomize' must be specified")
	}
	if hasManifest && hasKustomize {
		return errors.New("only one of 'manifest' or 'kustomize' can be specified")
	}

	// Validate kustomize.path is set when kustomize is used
	if c.Kustomize != nil && c.Kustomize.Path == "" {
		return errors.New("kustomize.path is required")
	}

	// Validate VCS config: required for kustomize, not allowed for inline manifest
	if c.Kustomize != nil {
		if err := c.basicVCSConfigRequest.Validate(); err != nil {
			return errors.Wrap(err, "kustomize requires a git source")
		}
	} else {
		// Inline manifest should not have VCS config
		if c.PublicGitVCSConfig != nil || c.ConnectedGithubVCSConfig != nil {
			return errors.New("VCS config is only valid for kustomize sources, not inline manifests")
		}
	}

	return nil
}

// @ID						CreateAppKubernetesManifestComponentConfig
// @Summary					create a kubernetes manifest component config
// @Description.markdown	create_kubernetes_manifest_component_config.md
// @Param					req				body	CreateKubernetesManifestComponentConfigRequest	true	"Input"
// @Param					app_id		path	string							true	"app ID"
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

	// Build VCS configs for kustomize sources
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid connected github vcs config: %w", err)
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid public vcs config: %w", err)
	}

	// build component config
	cfg := app.KubernetesManifestComponentConfig{
		Manifest:                 req.Manifest, // Empty for kustomize sources
		Namespace:                req.Namespace,
		PublicGitVCSConfig:       publicGitVCSConfig,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
	}

	// Populate kustomize config (mutually exclusive with Manifest)
	if req.Kustomize != nil {
		cfg.Kustomize = &app.KustomizeConfig{
			Path:           req.Kustomize.Path,
			Patches:        req.Kustomize.Patches,
			EnableHelm:     req.Kustomize.EnableHelm,
			LoadRestrictor: req.Kustomize.LoadRestrictor,
		}
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
