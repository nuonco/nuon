package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// RunnerComponentProbe is one synthetic health check the runner's
// component-health engine executes; Command is argv, never a shell string.
type RunnerComponentProbe struct {
	Type    string   `json:"type,omitempty"`
	Name    string   `json:"name,omitempty"`
	URL     string   `json:"url,omitempty"`
	Command []string `json:"command,omitempty"`
}

// RunnerInstallComponent is component metadata for the runner's component-health
// ownership index. It intentionally carries no credentials or cluster access.
type RunnerInstallComponent struct {
	InstallComponentID string `json:"install_component_id"`
	ComponentID        string `json:"component_id"`
	ComponentName      string `json:"component_name"`
	ComponentType      string `json:"component_type"`
	// HelmReleaseName and HelmNamespace map helm-managed cluster resources (via
	// the meta.helm.sh/release-name annotations) back to this component; helm-only.
	HelmReleaseName string                 `json:"helm_release_name,omitempty"`
	HelmNamespace   string                 `json:"helm_namespace,omitempty"`
	Probes          []RunnerComponentProbe `json:"probes,omitempty"`
}

type RunnerInstallComponentsResponse struct {
	InstallID  string                   `json:"install_id"`
	Components []RunnerInstallComponent `json:"components"`
}

// @ID						GetRunnerInstallComponents
// @Summary				list the install components a runner serves
// @Description			Returns metadata (IDs, name, type, declared health probes) for the components of the install this runner belongs to, used by the runner component-health engine to build its ownership index. Returns no credentials or cluster access.
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
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

	// Probes are vendor-declared commands and requests. Without the feature the
	// runner must not be asked to execute them.
	if enabled, _ := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureComponentHealth); !enabled {
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

	configs := s.resolveComponentConfigs(ctx, installID, installComponents)

	out := make([]RunnerInstallComponent, 0, len(installComponents))
	for _, ic := range installComponents {
		rc := RunnerInstallComponent{
			InstallComponentID: ic.ID,
			ComponentID:        ic.ComponentID,
			ComponentName:      ic.Component.Name,
			ComponentType:      string(ic.Component.Type),
		}
		if ccc, ok := configs[ic.ComponentID]; ok {
			if cfg := ccc.HelmComponentConfig; cfg != nil {
				rc.HelmReleaseName = cfg.ChartName
				rc.HelmNamespace = cfg.Namespace.ValueString()
			}
			// Never probe before first deploy — a pass would falsely report healthy.
			// Redeploys still probe: the previous workload is still serving.
			if ic.Status != app.InstallComponentStatusDisabled && ic.EverDeployed() {
				rc.Probes = runnerComponentProbes(ccc.HealthProbes)
			}
		}
		out = append(out, rc)
	}

	s.renderProbeTargets(ctx, installID, out)

	return &RunnerInstallComponentsResponse{
		InstallID:  installID,
		Components: out,
	}, nil
}

func runnerComponentProbes(probes app.ComponentHealthProbes) []RunnerComponentProbe {
	if len(probes) == 0 {
		return nil
	}

	out := make([]RunnerComponentProbe, 0, len(probes))
	for _, probe := range probes {
		out = append(out, RunnerComponentProbe{
			Type:    probe.Type,
			Name:    probe.Name,
			URL:     probe.URL,
			Command: probe.Command,
		})
	}
	return out
}

// renderProbeTargets interpolates install state into probe targets in place.
// Unresolved probes keep their template and surface as unknown, not dropped.
func (s *service) renderProbeTargets(ctx context.Context, installID string, comps []RunnerInstallComponent) {
	if !anyTemplatedProbe(comps) {
		return
	}

	installState, err := s.installsHelpers.GetInstallState(ctx, installID, false, true)
	if err != nil {
		s.l.Warn("unable to get install state to render health probe targets",
			zap.String("install_id", installID), zap.Error(err))
		return
	}
	stateData, err := installState.AsMap()
	if err != nil {
		s.l.Warn("unable to build install state map to render health probe targets",
			zap.String("install_id", installID), zap.Error(err))
		return
	}

	for i := range comps {
		for j, probe := range comps[i].Probes {
			rendered, err := renderProbe(probe, stateData)
			if err != nil {
				s.l.Warn("unable to render health probe target",
					zap.String("install_id", installID),
					zap.String("component_id", comps[i].ComponentID),
					zap.Error(err))
				continue
			}
			comps[i].Probes[j] = rendered
		}
	}
}

func renderProbe(probe RunnerComponentProbe, stateData map[string]any) (RunnerComponentProbe, error) {
	if probe.URL != "" {
		url, err := render.RenderV2(probe.URL, stateData)
		if err != nil {
			return probe, fmt.Errorf("unable to render probe url: %w", err)
		}
		probe.URL = url
	}

	if len(probe.Command) > 0 {
		// Clone first: probe.Command aliases the source config's backing array;
		// writing through it would mutate that config on a later render failure.
		command := slices.Clone(probe.Command)
		for idx, arg := range command {
			rendered, err := render.RenderV2(arg, stateData)
			if err != nil {
				return probe, fmt.Errorf("unable to render probe command: %w", err)
			}
			command[idx] = rendered
		}
		probe.Command = command
	}

	return probe, nil
}

func anyTemplatedProbe(comps []RunnerInstallComponent) bool {
	for _, comp := range comps {
		for _, probe := range comp.Probes {
			if probeIsTemplated(probe) {
				return true
			}
		}
	}
	return false
}

func probeIsTemplated(probe RunnerComponentProbe) bool {
	if strings.Contains(probe.URL, "{{") {
		return true
	}
	for _, arg := range probe.Command {
		if strings.Contains(arg, "{{") {
			return true
		}
	}
	return false
}

// resolveComponentConfigs maps component ID to config connection, falling back
// to the latest-configs view when a no-op sync reused connections. Best-effort.
func (s *service) resolveComponentConfigs(ctx context.Context, installID string, comps []app.InstallComponent) map[string]*app.ComponentConfigConnection {
	out := map[string]*app.ComponentConfigConnection{}

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
		out[cccs[i].ComponentID] = &cccs[i]
	}

	var missing []string
	for _, c := range comps {
		if !healthObservableComponentType(c.Component.Type) {
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
				if _, ok := out[fallback[i].ComponentID]; !ok {
					out[fallback[i].ComponentID] = &fallback[i]
				}
			}
		}
	}

	return out
}

func healthObservableComponentType(t app.ComponentType) bool {
	return t == app.ComponentTypeHelmChart || t == app.ComponentTypeKubernetesManifest
}
