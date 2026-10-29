package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallComponentHealthChecks
// @Summary				list custom component health checks
// @Description			Returns the latest reported state of every custom health check for the component (provider "custom"), keyed by check name.
// @Param					install_id		path	string	true	"install ID"
// @Param					component_id	path	string	true	"component ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallComponentResourceState
// @Router					/v1/installs/{install_id}/components/{component_id}/health/checks [get]
func (s *service) GetInstallComponentHealthChecks(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	checks, err := s.getInstallComponentHealthChecks(ctx, org.ID, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component health checks: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, checks)
}

func (s *service) getInstallComponentHealthChecks(ctx context.Context, orgID, installID, componentID string) ([]app.InstallComponentResourceState, error) {
	ic, err := s.findInstallComponent(ctx, orgID, installID, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install component: %w", err)
	}

	checks := make([]app.InstallComponentResourceState, 0)
	if err := s.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(app.InstallComponentResourceStatesLatestView)).
		Where(app.InstallComponentResourceState{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: ic.ID,
		}).
		Where("provider IN ?", []string{customCheckProvider, resourceProviderProbe}).
		Where(app.LatestReportOnlySQL(), app.LatestReportOnlyArgs(orgID, installID)...).
		Order("name").
		Find(&checks).Error; err != nil {
		return nil, fmt.Errorf("unable to query component health checks: %w", err)
	}

	return s.withDeclaredChecks(ctx, orgID, installID, ic, checks)
}

// withDeclaredChecks lists declared checks that have never reported as unknown.
// Without it a required check is invisible until a deploy fails on it.
func (s *service) withDeclaredChecks(
	ctx context.Context,
	orgID, installID string,
	ic *app.InstallComponent,
	observed []app.InstallComponentResourceState,
) ([]app.InstallComponentResourceState, error) {
	ccc, err := s.currentComponentConfig(ctx, installID, ic.ComponentID)
	if err != nil {
		// Config is context for the list, never a reason to fail the read.
		s.l.Warn("unable to load component config for declared health checks",
			zap.String("install_id", installID), zap.Error(err))
		return observed, nil
	}
	if ccc == nil {
		return observed, nil
	}

	type declaredCheck struct{ name, provider, kind string }
	declared := make([]declaredCheck, 0, len(ccc.HealthRequiredChecks)+len(ccc.HealthProbes))
	for _, name := range ccc.HealthRequiredChecks {
		declared = append(declared, declaredCheck{name, customCheckProvider, "CustomCheck"})
	}
	for _, probe := range ccc.HealthProbes.Runnable() {
		declared = append(declared, declaredCheck{probe.Name, resourceProviderProbe, probeResourceKind(probe.Type)})
	}

	seen := make(map[string]bool, len(observed))
	for _, c := range observed {
		seen[c.Provider+"/"+c.Name] = true
	}

	for _, d := range declared {
		key := d.provider + "/" + d.name
		if d.name == "" || seen[key] {
			continue
		}
		seen[key] = true
		observed = append(observed, app.InstallComponentResourceState{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: ic.ID,
			ComponentID:        ic.ComponentID,
			Provider:           d.provider,
			Kind:               d.kind,
			Name:               d.name,
			Health:             string(app.InstallComponentHealthStatusUnknown),
			Message:            "declared but not reported yet",
		})
	}

	sort.Slice(observed, func(i, j int) bool { return observed[i].Name < observed[j].Name })
	return observed, nil
}

func probeResourceKind(probeType string) string {
	switch strings.ToLower(strings.TrimSpace(probeType)) {
	case "tcp":
		return "TCPProbe"
	default:
		return "HTTPProbe"
	}
}

func (s *service) currentComponentConfig(ctx context.Context, installID, componentID string) (*app.ComponentConfigConnection, error) {
	var appConfigID string
	if err := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("app_config_id").
		Where("id = ?", installID).
		Scan(&appConfigID).Error; err != nil {
		return nil, err
	}

	var ccc app.ComponentConfigConnection
	if appConfigID != "" {
		err := s.db.WithContext(ctx).
			Where(app.ComponentConfigConnection{AppConfigID: appConfigID, ComponentID: componentID}).
			First(&ccc).Error
		if err == nil {
			return &ccc, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// An app config version only carries ccc rows for components CHANGED in that
	// sync, so the pin alone misses unchanged components — the same fallback the
	// deploy gate and probe handout use. Without it a component's declared checks
	// vanish from this list instead of reporting as not-yet-reported.
	if err := s.db.WithContext(ctx).
		Scopes(
			scopes.WithDisableViews,
			scopes.WithOverrideTable(app.LatestComponentConfigConnectionsViewName),
		).
		Where(app.ComponentConfigConnection{ComponentID: componentID}).
		First(&ccc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ccc, nil
}
