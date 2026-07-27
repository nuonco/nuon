package service

import (
	"context"
	"fmt"
	"net/http"

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

	return resources, nil
}
