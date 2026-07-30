package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

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

	return checks, nil
}
