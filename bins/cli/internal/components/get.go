package components

import (
	"context"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, appID, compID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	component, err := s.api.GetComponent(ctx, compID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(component)
		return
	}

	view.Render([][]string{
		{"id", component.ID},
		{"name", component.Name},
		{"created at", component.CreatedAt},
		{"updated at", component.UpdatedAt},
		{"created by", component.CreatedByID},
		{"app id ", component.AppID},
		{"config versions", strconv.Itoa(int(component.ConfigVersions))},
	})
}
