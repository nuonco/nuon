package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateHelmComponentConfigRequest struct {
	basicVCSConfigRequest
	HelmRepoConfig *HelmRepoConfigRequest `json:"helm_repo_config,omitempty"`

	Values                       map[string]*string `json:"values,omitempty" validate:"required"`
	ValuesFiles                  []string           `json:"values_files,omitempty"`
	ChartName                    string             `json:"chart_name,omitempty" validate:"required,dns_rfc1035_label,min=5,max=62"`
	Namespace                    string             `json:"namespace,omitempty"`
	StorageDriver                string             `json:"storage_driver,omitempty"`
	TakeOwnership                bool               `json:"take_ownership,omitempty"`
	SkipCRDs                     bool               `json:"skip_crds,omitempty"`
	BuildTimeout                 string             `json:"build_timeout,omitempty"`  // Duration string for build operations (e.g., "30m", "1h")
	DeployTimeout                string             `json:"deploy_timeout,omitempty"` // Duration string for deploy operations (e.g., "30m", "1h")
	MaxAutoRetries               *int               `json:"max_auto_retries,omitempty"`
	SkipNoops                    *bool              `json:"skip_noops,omitempty"`
	Toggleable                   *bool              `json:"toggleable,omitempty"`
	DefaultEnabled               *bool              `json:"default_enabled,omitempty"`
	AutoApproveOnPoliciesPassing *bool              `json:"auto_approve_on_policies_passing,omitempty"`
	HealthEnabled                *bool              `json:"health_enabled,omitempty" swaggertype:"boolean" extensions:"x-nullable"`
	HealthStabilizationWindow    string             `json:"health_stabilization_window,omitempty"` // Duration string for the health stabilization window (e.g., "3m")
	HealthBlockDeploy            *bool              `json:"health_block_deploy,omitempty" swaggertype:"boolean" extensions:"x-nullable"`

	HealthProbes         []HealthProbeRequest `json:"health_probes,omitempty"`
	HealthRequiredChecks []string             `json:"health_required_checks,omitempty"`

	AppConfigID string `json:"app_config_id"`

	Dependencies      []string                      `json:"dependencies"`
	References        []string                      `json:"references"`
	Checksum          string                        `json:"checksum"`
	DriftSchedule     *string                       `json:"drift_schedule,omitempty" validate:"omitempty,cron_schedule"`
	OperationRoles    map[app.OperationType]*string `json:"operation_roles,omitempty"`
	KubernetesContext string                        `json:"kubernetes_context,omitempty"`
}

type HelmRepoConfigRequest struct {
	RepoURL string `json:"repo_url" validate:"required,url"`
	Chart   string `json:"chart" validate:"required"`
	Version string `json:"version,omitempty"`
}

func (c *CreateHelmComponentConfigRequest) toConfig() *config.HelmChartComponentConfig {
	valuesFiles := make([]config.HelmValuesFile, 0, len(c.ValuesFiles))
	for _, contents := range c.ValuesFiles {
		valuesFiles = append(valuesFiles, config.HelmValuesFile{Contents: contents})
	}

	obj := &config.HelmChartComponentConfig{
		ChartName:     c.ChartName,
		Namespace:     c.Namespace,
		StorageDriver: c.StorageDriver,
		TakeOwnership: c.TakeOwnership,
		SkipCRDs:      c.SkipCRDs,
		ValuesMap:     build.DerefMap(c.Values),
		ValuesFiles:   valuesFiles,
	}
	if c.HelmRepoConfig != nil {
		obj.HelmRepo = &config.HelmRepoConfig{
			RepoURL: c.HelmRepoConfig.RepoURL,
			Chart:   c.HelmRepoConfig.Chart,
			Version: c.HelmRepoConfig.Version,
		}
	}
	return obj
}

func (c *CreateHelmComponentConfigRequest) buildInput(componentID string, depIDs []string) build.ComponentConnectionInput {
	return build.ComponentConnectionInput{
		ComponentID:                  componentID,
		AppConfigID:                  c.AppConfigID,
		References:                   c.References,
		Checksum:                     c.Checksum,
		DependencyIDs:                depIDs,
		BuildTimeout:                 c.BuildTimeout,
		DeployTimeout:                c.DeployTimeout,
		MaxAutoRetries:               c.MaxAutoRetries,
		SkipNoops:                    c.SkipNoops,
		Toggleable:                   c.Toggleable,
		DefaultEnabled:               c.DefaultEnabled,
		AutoApproveOnPoliciesPassing: c.AutoApproveOnPoliciesPassing,
		DriftSchedule:                c.DriftSchedule,
		HealthEnabled:                c.HealthEnabled,
		HealthStabilizationWindow:    c.HealthStabilizationWindow,
		HealthBlockDeploy:            c.HealthBlockDeploy,
		HealthProbes:                 toConfigHealthProbes(c.HealthProbes),
		HealthRequiredChecks:         c.HealthRequiredChecks,
		OperationRoles:               toConfigOperationRoles(c.OperationRoles),
		KubernetesContextName:        c.KubernetesContext,
	}
}

func (c *CreateHelmComponentConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.OperationRoles != nil {
		for operation := range c.OperationRoles {
			if !slices.Contains(app.ValidOperations, operation) {
				return fmt.Errorf("invalid operation type: %s. Valid operations: %v", operation, app.ValidOperations)
			}
		}
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		// Allow helm components without VCS config when using helm_repo_config
		if c.HelmRepoConfig != nil {
			if userErr, ok := err.(stderr.ErrUser); ok && userErr.Code == "vcs_config_required" {
				return nil
			}
		}
		return err
	}

	// Validate timeouts if provided
	if c.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(c.BuildTimeout); err != nil {
			return err
		}
	}
	if c.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(c.DeployTimeout); err != nil {
			return err
		}
	}
	if c.MaxAutoRetries != nil {
		if err := validation.ValidateMaxAutoRetries(*c.MaxAutoRetries); err != nil {
			return err
		}
	}
	if c.HealthStabilizationWindow != "" {
		if err := validation.ValidateHealthStabilizationWindow(c.HealthStabilizationWindow); err != nil {
			return err
		}
	}
	if err := validation.ValidateHealthProbeList(toConfigHealthProbes(c.HealthProbes)); err != nil {
		return err
	}
	if err := validation.ValidateRequiredChecks(c.HealthRequiredChecks); err != nil {
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
// @Failure				409	{object}	stderr.ErrResponse
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
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.HelmComponentConfig
// @Router					/v1/components/{component_id}/configs/helm [POST]
func (s *service) CreateHelmComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateHelmComponentConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	cfg, err := s.createHelmComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	if err := s.onConfigCreated(ctx, cmpID, app.ComponentTypeHelmChart); err != nil {
		ctx.Error(err)
		return
	}

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

	vcs := build.VCS{Github: connectedGithubVCSConfig, Public: publicGitVCSConfig}

	cfg, err := build.HelmComponentConfig(req.toConfig(), vcs)
	if err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}

	componentConfigConnection, err := build.ComponentConnection(req.buildInput(parentCmp.ID, depIDs))
	if err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	componentConfigConnection.HelmComponentConfig = cfg

	if res := s.db.WithContext(ctx).Create(componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create helm component config connection: %w", res.Error)
	}

	return cfg, nil
}
