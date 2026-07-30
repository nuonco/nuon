package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// ComponentHealthIncidentBundle is input for an agent or runbook: the most
// recent bad-verdict transition plus the component's currently non-healthy resources.
type ComponentHealthIncidentBundle struct {
	InstallComponentID string                              `json:"install_component_id"`
	CurrentHealth      string                              `json:"current_health"`
	Resolved           bool                                `json:"resolved"`
	Transition         HealthTransitionResponse            `json:"transition"`
	Resources          []app.InstallComponentResourceState `json:"resources"`
}

// @ID						GetInstallComponentHealthIncident
// @Summary				component health incident bundle
// @Description			Returns the most recent degraded/unhealthy transition for the component (whether or not it has since recovered) along with its diagnosis, correlated deploy, and the component's currently non-healthy resources. Returns a null body when there's no incident in the retained history. Requires the component-health feature.
// @Param					install_id				path	string	true	"install ID"
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
// @Success				200	{object}	service.ComponentHealthIncidentBundle
// @Router					/v1/installs/{install_id}/components/{component_id}/health/incident [get]
func (s *service) GetInstallComponentHealthIncident(ctx *gin.Context) {
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

	bundle, err := s.getInstallComponentHealthIncident(ctx, org.ID, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component health incident: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, bundle)
}

func (s *service) getInstallComponentHealthIncident(ctx context.Context, orgID, installID, componentID string) (*ComponentHealthIncidentBundle, error) {
	ic, err := s.findInstallComponent(ctx, orgID, installID, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install component: %w", err)
	}
	installComponentID := ic.ID

	t, err := s.findLatestBadTransition(ctx, orgID, installID, installComponentID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}

	resources, err := s.nonHealthyResources(ctx, orgID, installID, installComponentID)
	if err != nil {
		return nil, err
	}

	return &ComponentHealthIncidentBundle{
		InstallComponentID: installComponentID,
		CurrentHealth:      string(ic.HealthStatus),
		Resolved:           !ic.HealthStatus.IsBadHealth(),
		Transition: HealthTransitionResponse{
			FromHealth:            t.FromHealth,
			ToHealth:              t.ToHealth,
			Message:               t.Message,
			RootResourceKind:      t.RootResourceKind,
			RootResourceNamespace: t.RootResourceNamespace,
			RootResourceName:      t.RootResourceName,
			CorrelatedDeployID:    t.CorrelatedDeployID,
			Diagnosis:             t.Diagnosis,
			ObservedAt:            t.ObservedAt,
		},
		Resources: resources,
	}, nil
}
