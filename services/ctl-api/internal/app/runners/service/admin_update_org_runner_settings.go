package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/patcher"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type AdminUpdateOrgRunnerSettingsRequest struct {
	ContainerImageURL string `json:"container_image_url"`
	ContainerImageTag string `json:"container_image_tag"`
	RunnerAPIURL      string `json:"runner_api_url"`
	BinaryVersion     string `json:"binary_version"`
	RunnerBinaryURL   string `json:"runner_binary_url"`
}

// @ID						AdminUpdateOrgRunnerSettings
// @Summary				update runner settings for all runners in an org
// @Description			Updates container_image_tag and/or binary_version on every runner in the org and restarts them.
// @Param					org_id	path	string									true	"org ID"
// @Param					req		body	AdminUpdateOrgRunnerSettingsRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{array}		app.RunnerGroupSettings
// @Router					/v1/orgs/{org_id}/runner-settings [PATCH]
func (s *service) AdminUpdateOrgRunnerSettings(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminUpdateOrgRunnerSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if req.ContainerImageURL == "" && req.ContainerImageTag == "" && req.RunnerAPIURL == "" && req.BinaryVersion == "" && req.RunnerBinaryURL == "" {
		ctx.Error(fmt.Errorf("at least one field is required"))
		return
	}

	results, err := s.adminUpdateOrgRunnerSettings(ctx, orgID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update org runner settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (s *service) adminUpdateOrgRunnerSettings(ctx context.Context, orgID string, req *AdminUpdateOrgRunnerSettingsRequest) ([]*app.RunnerGroupSettings, error) {
	var runners []app.Runner
	res := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		Where(app.Runner{OrgID: orgID}).
		Find(&runners)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runners for org: %w", res.Error)
	}

	if len(runners) == 0 {
		return nil, fmt.Errorf("no runners found for org %s", orgID)
	}

	updates := app.RunnerGroupSettings{
		ContainerImageURL: req.ContainerImageURL,
		ContainerImageTag: req.ContainerImageTag,
		RunnerAPIURL:      req.RunnerAPIURL,
		BinaryVersion:     req.BinaryVersion,
		RunnerBinaryURL:   req.RunnerBinaryURL,
	}

	var results []*app.RunnerGroupSettings
	for _, runner := range runners {
		obj := app.RunnerGroupSettings{
			RunnerGroupID: runner.RunnerGroupID,
		}

		if res := s.db.WithContext(ctx).
			Scopes(scopes.WithPatcher(patcher.PatcherOptions{})).
			Where(obj).
			Updates(updates); res.Error != nil {
			return nil, fmt.Errorf("unable to update settings for runner %s: %w", runner.ID, res.Error)
		}

		results = append(results, &obj)

		s.l.Info("updated runner settings",
			zap.String("runner_id", runner.ID),
			zap.String("org_id", orgID),
			zap.String("container_image_tag", req.ContainerImageTag),
			zap.String("binary_version", req.BinaryVersion),
		)
	}

	return results, nil
}
