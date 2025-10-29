package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"gorm.io/gorm"
)

type GetComponentBuildRequest struct {
	ComponentBuildID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentBuildID
func (a *Activities) GetComponentBuild(ctx context.Context, req GetComponentBuildRequest) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	res := a.db.WithContext(ctx).
		Where("id = ?", req.ComponentBuildID).

		// load component config connection
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		Preload("ComponentConfigConnection.TerraformModuleComponentConfig").
		Preload("ComponentConfigConnection.HelmComponentConfig").
		Preload("ComponentConfigConnection.DockerBuildComponentConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig").
		Preload("ComponentConfigConnection.JobComponentConfig").
		Preload("ComponentConfigConnection.KubernetesManifestComponentConfig").
		First(&build)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, generics.TemporalGormError(gorm.ErrRecordNotFound, "component build not found")
		}
		return nil, fmt.Errorf("unable to load component build: %w", res.Error)
	}

	return &build, nil
}
