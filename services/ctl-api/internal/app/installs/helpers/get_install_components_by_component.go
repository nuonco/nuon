package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// getInstallComponents reads components deployed to an install from the DB.
func (h *Helpers) GetInstallComponentsByComponentID(ctx context.Context, installID string) (map[string]app.InstallComponent, error) {
	var components []app.InstallComponent
	res := h.db.WithContext(ctx).
		Where(&app.InstallComponent{InstallID: installID}).
		Find(&components)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	componentMap := make(map[string]app.InstallComponent, len(components))
	for _, component := range components {
		componentMap[component.ComponentID] = component
	}

	return componentMap, nil
}
