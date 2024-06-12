package components

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appNameOrID string, asJSON bool) {
	view := ui.NewListView()

	var (
		components []*models.AppComponent
		err        error
	)
	if appNameOrID != "" {
		appID, err := lookup.AppID(ctx, s.api, appNameOrID)
		if err != nil {
			view.Error(err)
			return
		}
		components, err = s.api.GetAppComponents(ctx, appID)
	} else {
		components, err = s.api.GetAllComponents(ctx)
	}
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(components)
		return
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"CREATED AT",
			"UPDATED AT",
			"CREATED BY",
			"CONFIG VERSIONS",
		},
	}
	for _, component := range components {
		data = append(data, []string{
			component.ID,
			component.Name,
			component.CreatedAt,
			component.UpdatedAt,
			component.CreatedByID,
			strconv.Itoa(int(component.ConfigVersions)),
		})
	}
	view.Render(data)
}
