package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetInstallComponentsBatchRequest struct {
	InstallID    string   `validate:"required"`
	ComponentIDs []string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallComponentsBatch(ctx context.Context, req GetInstallComponentsBatchRequest) (map[string]*app.InstallComponent, error) {
	var components []app.InstallComponent
	res := a.db.WithContext(ctx).
		Preload("Component").
		Preload("Component.Dependencies").
		Preload("TerraformWorkspace").
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC").Limit(1)
		}).
		Where("install_id = ? AND component_id IN ?", req.InstallID, req.ComponentIDs).
		Find(&components)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	result := make(map[string]*app.InstallComponent, len(components))
	for i := range components {
		result[components[i].ComponentID] = &components[i]
	}
	return result, nil
}
