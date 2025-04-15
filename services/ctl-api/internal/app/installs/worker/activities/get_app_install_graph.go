package activities

import (
	"context"
	"fmt"
	"slices"

	"github.com/dominikbraun/graph"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetAppInstallGraphRequest struct {
	AppID     string `json:"app_id"`
	InstallID string `json:"install_id"`

	Reverse bool `json:"reverse"`
}

// @temporal-gen activity
func (a *Activities) GetAppInstallGraph(ctx context.Context, req GetAppInstallGraphRequest) ([]string, error) {
	g, rootIDs, err := a.appsHelpers.GetGraph(ctx, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}

	componentIDs := make([]string, 0)
	for _, rootID := range rootIDs {
		if err := graph.BFS(g, rootID, func(compID string) bool {
			componentIDs = append(componentIDs, compID)
			return false
		}); err != nil {
			return nil, fmt.Errorf("unable to build app graph: %w", err)
		}
	}

	// remove install components that no longer exist but retain
	// the order of the componentIDs
	sanitizedCompIDs := make([]string, 0)

	for _, componentID := range componentIDs {
		installCmp := app.InstallComponent{}
		res := a.db.WithContext(ctx).
			Where(&app.InstallComponent{
				InstallID:   req.InstallID,
				ComponentID: componentID,
			}).
			First(&installCmp)
		if res.Error != nil && res.Error == gorm.ErrRecordNotFound {
			continue
		}

		if res.Error != nil {
			return nil, fmt.Errorf("unable to get install component: %w", res.Error)
		}

		if installCmp.Status == app.InstallComponentStatusDeleted {
			continue
		}

		sanitizedCompIDs = append(sanitizedCompIDs, componentID)
	}

	if req.Reverse {
		slices.Reverse(sanitizedCompIDs)
	}

	return sanitizedCompIDs, nil
}
