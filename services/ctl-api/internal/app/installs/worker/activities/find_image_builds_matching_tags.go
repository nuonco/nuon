package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FindImageBuildsMatchingTagsRequest struct {
	InstallID string   `json:"install_id" validate:"required"`
	Tags      []string `json:"tags"`
}

// ImageTagBuild is an image-component build whose source tag equals an install
// input. Helm charts pin that input as the image tag, which can move ahead of
// the install's app config, so the build has to be synced before the chart deploys.
type ImageTagBuild struct {
	ComponentID        string `json:"component_id"`
	ComponentName      string `json:"component_name"`
	BuildID            string `json:"build_id"`
	InstallComponentID string `json:"install_component_id"`
}

// @temporal-gen-v2 activity
func (a *Activities) FindImageBuildsMatchingTags(ctx context.Context, req FindImageBuildsMatchingTagsRequest) ([]ImageTagBuild, error) {
	if len(req.Tags) == 0 {
		return nil, nil
	}

	var install app.Install
	if err := a.db.WithContext(ctx).Select("app_id").First(&install, "id = ?", req.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to load install: %w", err)
	}

	var matches []ImageTagBuild
	if err := a.db.WithContext(ctx).Raw(`
		SELECT components.id AS component_id,
		       components.name AS component_name,
		       builds.id AS build_id,
		       ic.id AS install_component_id
		FROM `+app.LatestComponentConfigConnectionsViewName+` ccc
		JOIN components ON components.id = ccc.component_id AND components.deleted_at = 0
		JOIN external_image_component_configs img
		  ON img.component_config_connection_id = ccc.id AND img.deleted_at = 0
		JOIN component_builds builds
		  ON builds.id = ccc.latest_build_id AND builds.deleted_at = 0
		JOIN install_components ic
		  ON ic.component_id = components.id AND ic.install_id = ? AND ic.deleted_at = 0
		WHERE components.app_id = ?
		  AND components.type = ?
		  AND ccc.deleted_at = 0
		  AND img.tag IN ?
		  AND COALESCE(NULLIF(builds.status_v2->>'status', ''), builds.status) = ?
		ORDER BY components.name, components.id
	`, req.InstallID, install.AppID, app.ComponentTypeExternalImage, req.Tags, app.ComponentBuildStatusActive).
		Scan(&matches).Error; err != nil {
		return nil, fmt.Errorf("unable to find image builds for tags: %w", err)
	}
	if len(matches) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match.InstallComponentID)
	}
	var deploys []app.InstallDeploy
	if err := a.db.WithContext(ctx).
		Select("install_component_id", "component_build_id", "status", "created_at").
		Where("install_component_id IN ?", ids).
		Order("created_at DESC").
		Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to load image sync deploys: %w", err)
	}
	landed := map[string]app.InstallDeploy{}
	for _, deploy := range deploys {
		if _, ok := landed[deploy.InstallComponentID]; ok {
			continue
		}
		landed[deploy.InstallComponentID] = deploy
	}

	out := make([]ImageTagBuild, 0, len(matches))
	for _, match := range matches {
		deploy, ok := landed[match.InstallComponentID]
		if ok && imageTagSyncLanded(deploy.Status, deploy.ComponentBuildID, match.BuildID) {
			continue
		}
		out = append(out, match)
	}
	return out, nil
}

func imageTagSyncLanded(status app.InstallDeployStatus, deployedBuildID, wantBuildID string) bool {
	if deployedBuildID == "" || deployedBuildID != wantBuildID {
		return false
	}
	switch status {
	case app.InstallDeployStatusError, app.InstallDeployStatusCancelled:
		return false
	default:
		return true
	}
}
