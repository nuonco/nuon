package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type RefreshInstallHealthClusterAccessRequest struct {
	// RoleName is the identity health should read the cluster through. Empty
	// means the maintenance role, the same default drift and action runs use.
	RoleName string `json:"role_name"`
}

type RefreshInstallHealthClusterAccessResponse struct {
	ClusterFound bool   `json:"cluster_found"`
	ClusterID    string `json:"cluster_id,omitzero"`
	RoleName     string `json:"role_name,omitzero"`
}

// @ID						RefreshInstallHealthClusterAccess
// @Summary				refresh the cluster access component health reads through
// @Description			Derives the install's cluster access from its current stack outputs and the chosen role, then stores it for the runner's health engine. Use when health reports unknown because the install has not been deployed since component health was enabled, or after the cluster's endpoint or role changed. The runner picks the refreshed access up within a minute. Requires the component-health feature.
// @Param					req			body	RefreshInstallHealthClusterAccessRequest	false	"Input"
// @Param					install_id	path	string										true	"install ID"
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
// @Success				200	{object}	service.RefreshInstallHealthClusterAccessResponse
// @Router					/v1/installs/{install_id}/health/cluster-access [post]
func (s *service) RefreshInstallHealthClusterAccess(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireComponentHealthFeature(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	// An empty body is valid: it means "use the default role".
	var req RefreshInstallHealthClusterAccessRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Error(stderr.NewInvalidRequest(err))
			return
		}
	}

	resp, err := s.refreshHealthClusterAccess(ctx, org.ID, installID, req.RoleName)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) refreshHealthClusterAccess(ctx context.Context, orgID, installID, roleName string) (*RefreshInstallHealthClusterAccessResponse, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	access, err := s.helpers.ResolveComponentHealthClusterAccess(ctx, installID, roleName)
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         err,
			Description: "Unable to build cluster access for this install. Check that the chosen role exists on the install stack.",
		}
	}
	if access == nil {
		return &RefreshInstallHealthClusterAccessResponse{}, nil
	}

	raw, err := json.Marshal(access.ClusterInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal cluster info: %w", err)
	}

	// Keep the sandbox releases the runner discovered; only access is derived.
	update := app.Install{
		ComponentHealthContext: app.ComponentHealthContext{
			ClusterInfoJSON:     string(raw),
			SandboxHelmReleases: install.ComponentHealthContext.SandboxHelmReleases,
		},
	}
	if err := s.db.WithContext(ctx).
		Model(&app.Install{ID: installID}).
		Select("component_health_context").
		Updates(update).Error; err != nil {
		return nil, fmt.Errorf("unable to update component health context: %w", err)
	}

	return &RefreshInstallHealthClusterAccessResponse{
		ClusterFound: true,
		ClusterID:    access.ClusterInfo.ID,
		RoleName:     access.RoleName,
	}, nil
}
