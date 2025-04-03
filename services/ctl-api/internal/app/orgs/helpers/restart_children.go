package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
)

func (h *Helpers) GetEventLoops(ctx context.Context, orgID string) ([]bulk.EventLoop, error) {
	var org app.Org

	if res := h.db.WithContext(ctx).
		// runner group
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").

		// apps
		Preload("Apps").
		Preload("Apps.Installs").
		Preload("Apps.Installs.RunnerGroup").
		Preload("Apps.Installs.RunnerGroup.Runners").

		// components
		Preload("Apps.Components").

		// action workflows
		Preload("Apps.ActionWorkflows").

		// get org
		First(&org, "id = ?", orgID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get org")
	}

	return org.EventLoops(), nil
}
