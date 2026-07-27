package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// RunnerInstallComponent is component metadata for the runner's component-health
// ownership index. It intentionally carries no credentials or cluster access.
type RunnerInstallComponent struct {
	InstallComponentID string `json:"install_component_id"`
	ComponentID        string `json:"component_id"`
	ComponentName      string `json:"component_name"`
	ComponentType      string `json:"component_type"`
	// HelmReleaseName and HelmNamespace let the component-health engine map
	// helm-managed cluster resources (via the meta.helm.sh/release-name and
	// meta.helm.sh/release-namespace annotations) back to this component. Set
	// only for helm components.
	HelmReleaseName string `json:"helm_release_name,omitempty"`
	HelmNamespace   string `json:"helm_namespace,omitempty"`
}

type RunnerInstallComponentsResponse struct {
	InstallID  string                   `json:"install_id"`
	Components []RunnerInstallComponent `json:"components"`
}

// @ID						GetRunnerInstallComponents
// @Summary				list the install components a runner serves
// @Description			Returns metadata (IDs, name, type) for the components of the install this runner belongs to, used by the runner component-health engine to build its ownership index. Returns no credentials or cluster access.
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	service.RunnerInstallComponentsResponse
// @Router					/v1/runners/{runner_id}/install-components [GET]
func (s *service) GetRunnerInstallComponents(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	resp, err := s.getRunnerInstallComponents(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getRunnerInstallComponents(ctx context.Context, runnerID string) (*RunnerInstallComponentsResponse, error) {
	var runner app.Runner
	if err := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		First(&runner, "id = ?", runnerID).Error; err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.OwnerType != plugins.TableName(s.db, app.Install{}) {
		return &RunnerInstallComponentsResponse{Components: []RunnerInstallComponent{}}, nil
	}
	installID := runner.RunnerGroup.OwnerID

	var installComponents []app.InstallComponent
	if err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: installID}).
		Find(&installComponents).Error; err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}

	helmConfigs := s.resolveHelmConfigs(ctx, installID, installComponents)

	out := make([]RunnerInstallComponent, 0, len(installComponents))
	for _, ic := range installComponents {
		rc := RunnerInstallComponent{
			InstallComponentID: ic.ID,
			ComponentID:        ic.ComponentID,
			ComponentName:      ic.Component.Name,
			ComponentType:      string(ic.Component.Type),
		}
		if cfg, ok := helmConfigs[ic.ComponentID]; ok {
			rc.HelmReleaseName = cfg.ChartName
			rc.HelmNamespace = cfg.Namespace.ValueString()
		}
		out = append(out, rc)
	}

	return &RunnerInstallComponentsResponse{
		InstallID:  installID,
		Components: out,
	}, nil
}

// resolveHelmConfigs returns the helm config (chart name + namespace) per
// component id for the install's current app config, falling back to the
// latest-configs view for components whose connections were reused by a no-op
// sync. Best-effort: on error it returns what it has so identity metadata still
// flows.
func (s *service) resolveHelmConfigs(ctx context.Context, installID string, comps []app.InstallComponent) map[string]*app.HelmComponentConfig {
	out := map[string]*app.HelmComponentConfig{}

	var appConfigID string
	if err := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("app_config_id").
		Where("id = ?", installID).
		Scan(&appConfigID).Error; err != nil || appConfigID == "" {
		return out
	}

	var cccs []app.ComponentConfigConnection
	if err := s.db.WithContext(ctx).
		Preload("HelmComponentConfig").
		Where("app_config_id = ?", appConfigID).
		Find(&cccs).Error; err != nil {
		return out
	}
	for i := range cccs {
		if cccs[i].HelmComponentConfig != nil {
			out[cccs[i].ComponentID] = cccs[i].HelmComponentConfig
		}
	}

	var missing []string
	for _, c := range comps {
		if c.Component.Type != app.ComponentTypeHelmChart {
			continue
		}
		if _, ok := out[c.ComponentID]; !ok {
			missing = append(missing, c.ComponentID)
		}
	}
	if len(missing) > 0 {
		var fallback []app.ComponentConfigConnection
		if err := s.db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
			Preload("HelmComponentConfig").
			Where("component_id IN ?", missing).
			Find(&fallback).Error; err == nil {
			for i := range fallback {
				if _, ok := out[fallback[i].ComponentID]; !ok && fallback[i].HelmComponentConfig != nil {
					out[fallback[i].ComponentID] = fallback[i].HelmComponentConfig
				}
			}
		}
	}

	return out
}
