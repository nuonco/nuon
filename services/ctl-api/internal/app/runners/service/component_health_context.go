package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// ComponentHealthContextRequest carries what the runner's component-health
// engine needs to rehydrate cluster access after a restart.
type ComponentHealthContextRequest struct {
	ClusterInfoJSON     string   `json:"cluster_info_json" validate:"required"`
	SandboxHelmReleases []string `json:"sandbox_helm_releases"`
}

type ComponentHealthContextResponse struct {
	ClusterInfoJSON     string   `json:"cluster_info_json"`
	SandboxHelmReleases []string `json:"sandbox_helm_releases"`
}

// @ID						PutComponentHealthContext
// @Summary				persist the component-health context for a runner's install
// @Description			Stores the cluster access info and sandbox-managed helm release names the runner's component-health engine needs to rehydrate after a restart. No-op if the runner's group isn't owned by an install.
// @Param					req			body	ComponentHealthContextRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner ID"
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
// @Success				200
// @Router					/v1/runners/{runner_id}/component-health-context [PUT]
func (s *service) PutComponentHealthContext(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req ComponentHealthContextRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := s.putComponentHealthContext(ctx, runnerID, req); err != nil {
		ctx.Error(fmt.Errorf("unable to put component health context: %w", err))
		return
	}

	ctx.Status(http.StatusOK)
}

func (s *service) putComponentHealthContext(ctx context.Context, runnerID string, req ComponentHealthContextRequest) error {
	if enabled, _ := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureComponentHealth); !enabled {
		return nil
	}

	installID, ok, err := s.resolveComponentHealthInstallID(ctx, runnerID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	update := app.Install{
		ComponentHealthContext: app.ComponentHealthContext{
			ClusterInfoJSON:     req.ClusterInfoJSON,
			SandboxHelmReleases: req.SandboxHelmReleases,
		},
	}

	if err := s.db.WithContext(ctx).
		Model(&app.Install{ID: installID}).
		Select("component_health_context").
		Updates(update).Error; err != nil {
		return fmt.Errorf("unable to update install component health context: %w", err)
	}

	return nil
}

// @ID						GetComponentHealthContext
// @Summary				get the component-health context for a runner's install
// @Description			Returns the cluster access info and sandbox-managed helm release names the runner's component-health engine needs to rehydrate after a restart. Returns an empty struct if unset, or if the runner's group isn't owned by an install.
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
// @Success				200	{object}	service.ComponentHealthContextResponse
// @Router					/v1/runners/{runner_id}/component-health-context [GET]
func (s *service) GetComponentHealthContext(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	resp, err := s.getComponentHealthContext(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component health context: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getComponentHealthContext(ctx context.Context, runnerID string) (*ComponentHealthContextResponse, error) {
	if enabled, _ := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureComponentHealth); !enabled {
		return &ComponentHealthContextResponse{}, nil
	}

	installID, ok, err := s.resolveComponentHealthInstallID(ctx, runnerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &ComponentHealthContextResponse{SandboxHelmReleases: []string{}}, nil
	}

	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	resp := &ComponentHealthContextResponse{
		ClusterInfoJSON:     install.ComponentHealthContext.ClusterInfoJSON,
		SandboxHelmReleases: install.ComponentHealthContext.SandboxHelmReleases,
	}
	if resp.SandboxHelmReleases == nil {
		resp.SandboxHelmReleases = []string{}
	}

	return resp, nil
}

// resolveComponentHealthInstallID resolves the install a runner belongs to,
// exactly like getRunnerInstallComponents. Returns ok=false (no error) when
// the runner's group isn't owned by an install.
func (s *service) resolveComponentHealthInstallID(ctx context.Context, runnerID string) (string, bool, error) {
	var runner app.Runner
	if err := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		Where(app.Runner{ID: runnerID}).
		First(&runner).Error; err != nil {
		return "", false, fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.OwnerType != plugins.TableName(s.db, app.Install{}) {
		return "", false, nil
	}

	return runner.RunnerGroup.OwnerID, true, nil
}
