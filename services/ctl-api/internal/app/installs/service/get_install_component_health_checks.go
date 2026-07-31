package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallComponentHealthChecks
// @Summary				list custom component health checks
// @Description			Returns the latest reported state of every custom health check for the component (provider "custom"), keyed by check name. Requires the component-health feature.
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
	if err := s.requireComponentHealthFeature(ctx, org); err != nil {
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
		Scopes(scopes.WithOverrideTable(views.CurrentViewName(s.chDB, &app.InstallComponentResourceState{}))).
		Where(app.InstallComponentResourceState{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: ic.ID,
			Provider:           customCheckProvider,
		}).
		Order("name").
		Find(&checks).Error; err != nil {
		return nil, fmt.Errorf("unable to query custom health checks: %w", err)
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

	declared := make([]string, 0, len(ccc.HealthRequiredChecks)+len(ccc.HealthProbes))
	declared = append(declared, ccc.HealthRequiredChecks...)
	for _, probe := range ccc.HealthProbes {
		declared = append(declared, probe.Name)
	}

	seen := make(map[string]bool, len(observed))
	for _, c := range observed {
		seen[c.Name] = true
	}

	for _, name := range declared {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		observed = append(observed, app.InstallComponentResourceState{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: ic.ID,
			ComponentID:        ic.ComponentID,
			Provider:           customCheckProvider,
			Kind:               "CustomCheck",
			Name:               name,
			Health:             string(app.InstallComponentHealthStatusUnknown),
			Message:            "declared but not reported yet",
		})
	}

	sort.Slice(observed, func(i, j int) bool { return observed[i].Name < observed[j].Name })
	return observed, nil
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
	if appConfigID == "" {
		return nil, nil
	}

	var ccc app.ComponentConfigConnection
	if err := s.db.WithContext(ctx).
		Where(app.ComponentConfigConnection{AppConfigID: appConfigID, ComponentID: componentID}).
		First(&ccc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ccc, nil
}
