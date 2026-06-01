package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetInstallComponentsByComponentIDsInput struct {
	InstallID    string   `json:"install_id" validate:"required"`
	ComponentIDs []string `json:"component_ids" validate:"required"`
}

type InstallComponentMapping struct {
	ComponentID        string `json:"component_id"`
	InstallComponentID string `json:"install_component_id"`
}

type GetInstallComponentsByComponentIDsOutput struct {
	Mappings []InstallComponentMapping `json:"mappings"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) getInstallComponentsByComponentIDs(ctx context.Context, req *GetInstallComponentsByComponentIDsInput) (*GetInstallComponentsByComponentIDsOutput, error) {
	var installComponents []app.InstallComponent
	err := a.db.WithContext(ctx).
		Where("install_id = ? AND component_id IN ?", req.InstallID, req.ComponentIDs).
		Find(&installComponents).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}

	mappings := make([]InstallComponentMapping, 0, len(installComponents))
	for _, ic := range installComponents {
		mappings = append(mappings, InstallComponentMapping{
			ComponentID:        ic.ComponentID,
			InstallComponentID: ic.ID,
		})
	}

	return &GetInstallComponentsByComponentIDsOutput{
		Mappings: mappings,
	}, nil
}
