package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

const defaultInstallResourcesLimit = 5_000

type installResourceFilters struct {
	InstallComponentID string
	Kind               string
	Namespace          string
	Health             string
	Provider           string
}

// @ID					GetInstallResources
// @Summary				live resource explorer for an install
// @Description			Returns the latest observed state of every resource the install's components manage, filterable by component, kind, namespace, health, and provider. Requires the component-health feature.
// @Param				install_id				path	string	true	"install ID"
// @Param				install_component_id	query	string	false	"filter by install component ID"
// @Param				kind					query	string	false	"filter by resource kind (e.g. Deployment)"
// @Param				namespace				query	string	false	"filter by namespace"
// @Param				health					query	string	false	"filter by health (healthy|progressing|degraded|unhealthy|unknown)"
// @Param				provider				query	string	false	"filter by provider (kubernetes|aws|gcp|azure)"
// @Tags				installs
// @Accept				json
// @Produce				json
// @Security			APIKey
// @Security			OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallComponentResourceState
// @Router				/v1/installs/{install_id}/resources [GET]
func (s *service) GetInstallResources(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureComponentHealth)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check component-health feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("component health is not enabled for org %s", org.ID),
			Description: "The component health feature is not enabled for this organization.",
		})
		return
	}

	resources, err := s.getInstallResources(ctx, org.ID, installID, installResourceFilters{
		InstallComponentID: ctx.Query("install_component_id"),
		Kind:               ctx.Query("kind"),
		Namespace:          ctx.Query("namespace"),
		Health:             ctx.Query("health"),
		Provider:           ctx.Query("provider"),
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install resources: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resources)
}

func (s *service) getInstallResources(ctx context.Context, orgID, installID string, f installResourceFilters) ([]app.InstallComponentResourceState, error) {
	resources := make([]app.InstallComponentResourceState, 0)

	q := s.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(views.CurrentViewName(s.chDB, &app.InstallComponentResourceState{}))).
		Where(app.InstallComponentResourceState{
			OrgID:     orgID,
			InstallID: installID,
		})

	if f.InstallComponentID != "" {
		q = q.Where(app.InstallComponentResourceState{InstallComponentID: f.InstallComponentID})
	}
	if f.Kind != "" {
		q = q.Where(app.InstallComponentResourceState{Kind: f.Kind})
	}
	if f.Namespace != "" {
		q = q.Where(app.InstallComponentResourceState{Namespace: f.Namespace})
	}
	if f.Health != "" {
		q = q.Where(app.InstallComponentResourceState{Health: f.Health})
	}
	if f.Provider != "" {
		q = q.Where(app.InstallComponentResourceState{Provider: f.Provider})
	}

	res := q.
		Order("install_component_id, kind, namespace, name").
		Limit(defaultInstallResourcesLimit).
		Find(&resources)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to query resource states: %w", res.Error)
	}

	deployed, err := s.deployedInstallComponentIDs(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	resources = retainDeployedComponentResources(resources, deployed)

	declared, err := s.declaredProbesByInstallComponent(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	markRemovedProbes(resources, declared)

	return resources, nil
}

// declaredProbesByInstallComponent maps each install component to the probe
// names its CURRENT config declares, from the install's pinned app config.
func (s *service) declaredProbesByInstallComponent(ctx context.Context, orgID, installID string) (map[string]map[string]bool, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Select("id", "app_config_id").
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	if install.AppConfigID == "" {
		return map[string]map[string]bool{}, nil
	}

	var cccs []app.ComponentConfigConnection
	if err := s.db.WithContext(ctx).
		Where(app.ComponentConfigConnection{AppConfigID: install.AppConfigID}).
		Find(&cccs).Error; err != nil {
		return nil, fmt.Errorf("unable to list component configs: %w", err)
	}
	byComponent := map[string]map[string]bool{}
	for i := range cccs {
		byComponent[cccs[i].ComponentID] = probeNameSet(cccs[i].HealthProbes)
	}

	var comps []app.InstallComponent
	if err := s.db.WithContext(ctx).
		Select("id", "component_id").
		Where(app.InstallComponent{InstallID: installID, OrgID: orgID}).
		Find(&comps).Error; err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}

	// An app config version only carries ccc rows for components CHANGED in
	// that sync. Components missing from the pinned version resolve via the
	// latest-configs view — the same fallback the runner probe handout uses —
	// otherwise a no-op sync marks every live probe as removed.
	var missing []string
	for i := range comps {
		if _, ok := byComponent[comps[i].ComponentID]; !ok {
			missing = append(missing, comps[i].ComponentID)
		}
	}
	if len(missing) > 0 {
		var fallback []app.ComponentConfigConnection
		if err := s.db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
			Where("component_id IN ?", missing).
			Find(&fallback).Error; err != nil {
			return nil, fmt.Errorf("unable to resolve current component configs: %w", err)
		}
		for i := range fallback {
			if _, ok := byComponent[fallback[i].ComponentID]; !ok {
				byComponent[fallback[i].ComponentID] = probeNameSet(fallback[i].HealthProbes)
			}
		}
	}

	out := map[string]map[string]bool{}
	for i := range comps {
		if names, ok := byComponent[comps[i].ComponentID]; ok {
			out[comps[i].ID] = names
		} else {
			out[comps[i].ID] = map[string]bool{}
		}
	}
	return out, nil
}

func probeNameSet(probes app.ComponentHealthProbes) map[string]bool {
	names := map[string]bool{}
	for _, probe := range probes {
		if probe.Name != "" {
			names[probe.Name] = true
		}
	}
	return names
}

// markRemovedProbes labels probe rows whose name is no longer declared in the
// owning component's current config. Observations persist in ClickHouse for
// days after a probe is deleted from config; without the label the last
// reading keeps rendering as a live check.
func markRemovedProbes(resources []app.InstallComponentResourceState, declared map[string]map[string]bool) {
	for i := range resources {
		r := &resources[i]
		if r.Source != app.InstallComponentResourceSourceComponent || !isProbeKind(r.Kind) {
			continue
		}
		names, known := declared[r.InstallComponentID]
		if known && !names[r.Name] {
			r.RemovedFromConfig = true
		}
	}
}

func isProbeKind(kind string) bool {
	return strings.HasSuffix(kind, "Probe")
}

func (s *service) deployedInstallComponentIDs(ctx context.Context, orgID, installID string) (map[string]bool, error) {
	var comps []app.InstallComponent
	if err := s.db.WithContext(ctx).
		Select("id", "status", "health_status").
		Where(app.InstallComponent{InstallID: installID, OrgID: orgID}).
		Find(&comps).Error; err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}

	deployed := make(map[string]bool, len(comps))
	for i := range comps {
		// EverDeployed, not HasDeployed: a redeploy transiently leaves the
		// deployed set (planning/syncing/executing), and hiding a component's
		// resources for the duration of every deploy is exactly when someone
		// is watching them. Torn-down components resolve false again.
		deployed[comps[i].ID] = comps[i].Status != app.InstallComponentStatusDisabled && comps[i].EverDeployed()
	}
	return deployed, nil
}

// retainDeployedComponentResources drops rows belonging to a component that has
// never deployed. Observations live in ClickHouse for days, so a row recorded
// before a component was torn down — or during the window where it briefly
// looked deployed — keeps rendering with a fresh timestamp long after there is
// any workload to observe. Showing them invites the reader to judge a component
// by resources that no longer describe anything.
//
// Sandbox rows are always kept: they belong to the install's own infrastructure,
// not to a component, so no component deploy state governs them.
func retainDeployedComponentResources(resources []app.InstallComponentResourceState, deployed map[string]bool) []app.InstallComponentResourceState {
	kept := make([]app.InstallComponentResourceState, 0, len(resources))
	for _, r := range resources {
		if r.Source == app.InstallComponentResourceSourceComponent && !deployed[r.InstallComponentID] {
			continue
		}
		kept = append(kept, r)
	}
	return kept
}
